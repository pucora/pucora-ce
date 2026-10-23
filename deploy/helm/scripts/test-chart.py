"""Render real Helm templates and check routing, isolation and validation.

Run: python -m pip install pyyaml && python deploy/helm/scripts/test-chart.py
"""
import json
from pathlib import Path
import subprocess
import unittest

import yaml

CHART = Path(__file__).resolve().parents[1] / "pucora"


def render(*args, valid=True):
    result = subprocess.run(
        ["helm", "template", "audit", str(CHART), "--namespace", "gateway", *args],
        capture_output=True, text=True, encoding="utf-8",
    )
    if not valid:
        if result.returncode == 0:
            raise AssertionError("Invalid configuration rendered successfully")
        return result.stderr
    if result.returncode:
        raise AssertionError(result.stderr)
    return [doc for doc in yaml.safe_load_all(result.stdout) if doc]


def resource(docs, kind, name="audit-pucora"):
    return next(d for d in docs if d["kind"] == kind and d["metadata"]["name"] == name)


class ChartTests(unittest.TestCase):
    def test_examples(self):
        for path in (CHART / "ci").glob("*.yaml"):
            with self.subTest(values=path.name):
                render("-f", str(path))

    def test_gateway_isolated_from_configurator(self):
        docs = render("--set", "configurator.enabled=true,podMonitor.enabled=true,podDisruptionBudget.enabled=true,networkPolicy.enabled=true")
        gateway = resource(docs, "Deployment")
        selector = gateway["spec"]["selector"]["matchLabels"]
        self.assertEqual(selector["app.kubernetes.io/component"], "gateway")
        for obj in docs:
            if obj["kind"] == "Deployment":
                labels = obj["spec"]["template"]["metadata"]["labels"]
                matches = all(labels.get(k) == v for k, v in selector.items())
                self.assertEqual(matches, obj is gateway)
        for name in ("audit-pucora", "audit-pucora-metrics"):
            self.assertEqual(resource(docs, "Service", name)["spec"]["selector"], selector)
        for kind in ("PodMonitor", "PodDisruptionBudget"):
            self.assertEqual(resource(docs, kind)["spec"]["selector"]["matchLabels"], selector)
        self.assertEqual(resource(docs, "NetworkPolicy")["spec"]["podSelector"]["matchLabels"], selector)

    def test_network_namespaces(self):
        docs = render("--set", "networkPolicy.enabled=true,networkPolicy.ingressControllerNamespace=ingress-nginx,networkPolicy.metricsNamespace=monitoring")
        rules = resource(docs, "NetworkPolicy")["spec"]["ingress"]
        for rule, namespace in zip(rules, ["ingress-nginx", "monitoring"]):
            self.assertEqual(rule["from"], [{"namespaceSelector": {"matchLabels": {"kubernetes.io/metadata.name": namespace}}}])

    def test_network_empty_namespace_is_local(self):
        docs = render("--set", "networkPolicy.enabled=true")
        for rule in resource(docs, "NetworkPolicy")["spec"]["ingress"]:
            self.assertEqual(rule["from"][0]["namespaceSelector"]["matchLabels"]["kubernetes.io/metadata.name"], "gateway")

    def test_network_deny_all(self):
        docs = render("--set", "networkPolicy.enabled=true,networkPolicy.allowIngressController=false,networkPolicy.allowMetricsScraper=false")
        self.assertEqual(resource(docs, "NetworkPolicy")["spec"]["ingress"], [])

    def test_prometheus_default(self):
        docs = render("--set", "serviceMonitor.enabled=true")
        config = json.loads(resource(docs, "ConfigMap")["data"]["pucora.json"])
        exporter = config["extra_config"]["telemetry/opencensus"]["exporters"]["prometheus"]
        service = resource(docs, "Service", "audit-pucora-metrics")
        self.assertEqual(exporter["port"], service["spec"]["ports"][0]["port"])
        self.assertEqual(config["endpoints"], [])
        self.assertEqual(resource(docs, "ServiceMonitor")["spec"]["endpoints"][0]["path"], "/metrics")

    def test_port_override(self):
        docs = render("--set", "service.targetPort=8888")
        container = resource(docs, "Deployment")["spec"]["template"]["spec"]["containers"][0]
        self.assertEqual(container["args"][-1], "8888")
        self.assertEqual(container["ports"][0]["containerPort"], 8888)

    def test_metrics_port_and_disable(self):
        for settings, enabled in (("metrics.port=9099", True), ("metrics.enabled=false", False)):
            docs = render("--set", settings)
            config = json.loads(resource(docs, "ConfigMap")["data"]["pucora.json"])
            exporters = config["extra_config"]["telemetry/opencensus"]["exporters"]
            self.assertEqual("prometheus" in exporters, enabled)
            if enabled:
                self.assertEqual(exporters["prometheus"]["port"], 9099)
            else:
                self.assertFalse(any(d["kind"] == "Service" and d["metadata"]["name"].endswith("-metrics") for d in docs))

    def test_configurator_auth_health_and_persistence(self):
        docs = render("--set", "configurator.enabled=true,configurator.api.auth.existingSecret=api-key")
        api = resource(docs, "Deployment", "audit-pucora-configurator-api")
        self.assertEqual(api["spec"]["strategy"]["type"], "Recreate")
        container = api["spec"]["template"]["spec"]["containers"][0]
        key = next(e for e in container["env"] if e["name"] == "CONFIG_API_KEY")
        self.assertEqual(key["valueFrom"]["secretKeyRef"]["name"], "api-key")
        for key in ("livenessProbe", "readinessProbe"):
            self.assertEqual(container[key]["httpGet"]["path"], "/api/health")

    def test_configurator_fetch_auth_custom_key(self):
        docs = render("--set", "configurator.enabled=true,config.mode=configurator,configurator.api.auth.existingSecret=api-key,configurator.api.auth.secretKey=custom")
        pod = resource(docs, "Deployment")["spec"]["template"]["spec"]
        secret = next(v for v in pod["volumes"] if v["name"] == "configurator-api-key")["secret"]
        self.assertEqual(secret["items"][0]["key"], "custom")
        self.assertIn("X-API-Key:", pod["initContainers"][0]["command"][-1])

    def test_default_config_mounted_and_checked(self):
        docs = render()
        for pod in (resource(docs, "Deployment")["spec"]["template"]["spec"], resource(docs, "Pod", "audit-pucora-config-test")["spec"]):
            volume = next(v for v in pod["volumes"] if v["name"] == "config")
            self.assertEqual(volume["configMap"]["name"], "audit-pucora")
            self.assertTrue(any(v["mountPath"] == "/etc/pucora/pucora.json" for v in pod["containers"][0]["volumeMounts"]))

    def test_pdb_zero_and_max_unavailable(self):
        docs = render("--set", "podDisruptionBudget.enabled=true,podDisruptionBudget.minAvailable=0")
        self.assertEqual(resource(docs, "PodDisruptionBudget")["spec"]["minAvailable"], 0)
        docs = render("--set", "podDisruptionBudget.enabled=true,podDisruptionBudget.minAvailable=null,podDisruptionBudget.maxUnavailable=0")
        self.assertEqual(resource(docs, "PodDisruptionBudget")["spec"]["maxUnavailable"], 0)

    def test_security_and_test_image(self):
        docs = render("--set", "image.pullPolicy=Never,imagePullSecrets[0].name=private-registry")
        self.assertFalse(resource(docs, "Deployment")["spec"]["template"]["spec"]["automountServiceAccountToken"])
        pod = resource(docs, "Pod", "audit-pucora-config-test")["spec"]
        self.assertFalse(pod["automountServiceAccountToken"])
        self.assertEqual(pod["imagePullSecrets"], [{"name": "private-registry"}])
        self.assertEqual(pod["containers"][0]["imagePullPolicy"], "Never")
        self.assertTrue(pod["containers"][0]["securityContext"]["runAsNonRoot"])

    def test_invalid_settings_rejected(self):
        cases = {
            "autoscaling.enabled=true,keda.enabled=true": "mutually exclusive",
            "autoscaling.enabled=true,autoscaling.minReplicas=5,autoscaling.maxReplicas=2": "must not exceed",
            "autoscaling.enabled=true,autoscaling.targetCPUUtilizationPercentage=0": "requires a CPU",
            "autoscaling.enabled=true,resources.requests.cpu=null": "requests.cpu",
            "keda.enabled=true": "at least one trigger",
            "podDisruptionBudget.enabled=true,podDisruptionBudget.maxUnavailable=1": "exactly one",
            "configurator.enabled=true,configurator.api.replicaCount=2": "local file store",
            "config.mode=configurator": "requires configurator.enabled",
            "service.targetPort=70000": "65535",
        }
        for settings, message in cases.items():
            with self.subTest(settings=settings):
                self.assertIn(message, render("--set", settings, valid=False))

    def test_secret_and_external_config(self):
        docs = render("--set", "config.mode=secret,config.existingSecret=external")
        self.assertFalse(any(d["kind"] in ("ConfigMap", "Secret") for d in docs))
        pod = resource(docs, "Deployment")["spec"]["template"]["spec"]
        self.assertEqual(next(v for v in pod["volumes"] if v["name"] == "config")["secret"]["secretName"], "external")


if __name__ == "__main__":
    unittest.main(verbosity=2)
