# Pucora Helm Chart

Deploy the [Pucora API Gateway](https://pucora.io) on Kubernetes.

**Source repository:** [github.com/pucora/pucora-ce](https://github.com/pucora/pucora-ce)  
**Chart path:** [`deploy/helm/pucora/`](https://github.com/pucora/pucora-ce/tree/main/deploy/helm/pucora)

## Prerequisites

- Kubernetes 1.23+
- Helm 3.8+
- (Optional) Prometheus Operator for `ServiceMonitor` / `PodMonitor` / `PrometheusRule`
- (Optional) cert-manager for automatic TLS certificates
- (Optional) Gateway API controller for `HTTPRoute`
- (Optional) KEDA for event-driven autoscaling

## Quick start

```bash
git clone https://github.com/pucora/pucora-ce.git
cd pucora-ce
helm install my-gateway ./deploy/helm/pucora
```

Verify:

```bash
kubectl port-forward svc/my-gateway-pucora 8080:8080
curl http://localhost:8080/__health
helm test my-gateway
```

### Install from OCI registry (on release)

```bash
helm install my-gateway oci://ghcr.io/pucora/charts/pucora --version 2.1.1
```

## Versioning discipline

Chart metadata is kept in sync with the gateway release version in:

- `Makefile` → `VERSION`
- `deploy/helm/pucora/Chart.yaml` → `version` and `appVersion`
- `deploy/helm/pucora/values.yaml` → `image.tag`

Before tagging a release:

```bash
# Bump VERSION in Makefile, then:
make sync-chart-version
make verify-chart-version
```

CI enforces alignment on every PR (`go.yml`) and before chart publish (`release.yml`). Cluster install tests run in [`.github/workflows/helm-cluster.yml`](../../../.github/workflows/helm-cluster.yml).

Local cluster smoke test (requires Docker + Kind):

```bash
make helm-cluster-test
```

## Example values files

| File | Use case |
|------|----------|
| [`ci/values-prod.yaml`](ci/values-prod.yaml) | Production: image config, HPA, PDB, NetworkPolicy, monitoring |
| [`ci/values-aws-nlb.yaml`](ci/values-aws-nlb.yaml) | AWS NLB with cross-zone spread |
| [`ci/values-istio.yaml`](ci/values-istio.yaml) | Istio sidecar injection with startup probe |

```bash
helm install my-gateway ./deploy/helm/pucora -f deploy/helm/pucora/ci/values-prod.yaml
```

## Configuration modes

| Mode | Description |
|------|-------------|
| `configmap` (default) | Mount `pucora.json` from a ConfigMap |
| `secret` | Mount `pucora.json` from a Secret (sensitive credentials) |
| `image` | Config baked into custom Docker image (production best practice) |

```bash
# Custom config file
helm install my-gateway ./deploy/helm/pucora \
  --set-file config.pucoraJson=./pucora.json

# Secret mode
helm install my-gateway ./deploy/helm/pucora \
  --set config.mode=secret \
  --set-file config.pucoraJson=./pucora.json

# Immutable image
helm install my-gateway ./deploy/helm/pucora \
  --set config.mode=image \
  --set image.repository=myregistry/pucora-gateway \
  --set image.tag=v1.0.0
```

Validate config before deploy (recommended in CI/CD):

```bash
docker run --rm -v $(pwd)/pucora.json:/etc/pucora/pucora.json \
  niteesh20/pucora:2.1.1 check --lint -c /etc/pucora/pucora.json
```

## Production hardening

### Startup probe, graceful shutdown

```yaml
probes:
  startup:
    enabled: true

lifecycle:
  preStop:
    enabled: true
    sleepSeconds: 10

terminationGracePeriodSeconds: 30
```

Enable startup probe when using Istio/Linkerd sidecars.

### High availability

```yaml
podAntiAffinity:
  enabled: true
  type: hard

topologySpreadConstraints:
  - maxSkew: 1
    topologyKey: topology.kubernetes.io/zone
    whenUnsatisfiable: ScheduleAnyway
    labelSelector:
      matchLabels:
        app.kubernetes.io/name: pucora

podDisruptionBudget:
  enabled: true
  minAvailable: 1
```

### Extra volumes

```yaml
extraVolumes:
  - name: tls-certs
    secret:
      secretName: backend-ca
extraVolumeMounts:
  - name: tls-certs
    mountPath: /etc/ssl/certs/backend-ca.pem
    subPath: ca.pem
    readOnly: true
```

## Networking

### AWS NLB

```yaml
service:
  nlb:
    enabled: true
    externalTrafficPolicy: Local
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
      service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
      service.beta.kubernetes.io/aws-load-balancer-additional-resource-tags: "Environment=prod"
```

### Ingress + cert-manager

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix

certificate:
  enabled: true
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
```

### Gateway API (HTTPRoute)

```yaml
httpRoute:
  enabled: true
  parentRefs:
    - name: main-gateway
      namespace: gateway-system
  hostnames:
    - api.example.com
```

### NetworkPolicy

```yaml
networkPolicy:
  enabled: true
  allowIngressController: true
  allowMetricsScraper: true
```

## Sidecar injection

```yaml
sidecarInjection:
  enabled: true
  istio:
    enabled: true
    holdApplicationUntilProxyStarts: true

sidecars:
  - name: oauth2-proxy
    image: quay.io/oauth2-proxy/oauth2-proxy:v7.6.0
```

## Scaling

### HPA (CPU/memory)

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

### KEDA (request-rate, custom metrics)

```yaml
keda:
  enabled: true
  minReplicaCount: 2
  maxReplicaCount: 20
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus:9090
        query: sum(rate(http_requests_total{service="pucora"}[2m]))
        threshold: "100"
```

HPA and KEDA are mutually exclusive.

## Observability

```yaml
usageReporting:
  disable: true

opentelemetry:
  enabled: true
  endpoint: http://otel-collector:4317

serviceMonitor:
  enabled: true

podMonitor:
  enabled: true

prometheusRule:
  enabled: true
  rules:
    - alert: PucoraDown
      expr: up{job=~".*pucora.*"} == 0
      for: 5m
      labels:
        severity: critical
```

## Security

- Runs as UID **1000**, `readOnlyRootFilesystem`, drops all caps except `NET_BIND_SERVICE`
- `USAGE_DISABLE=true` by default
- Optional RBAC for ServiceAccount:

```yaml
rbac:
  create: true
  rules:
    - apiGroups: [""]
      resources: ["secrets"]
      verbs: ["get"]
```

## GitOps

```yaml
gitops:
  argocd:
    enabled: true
    syncWave: "2"
```

## Blue/green deployments

Pucora is stateless. For zero-downtime config changes:

1. Build a new image with updated `pucora.json` (`config.mode=image`)
2. Deploy a second Helm release (e.g. `my-gateway-green`)
3. Switch the Service selector or Ingress backend to the new release
4. Remove the old release

Alternatively use Argo Rollouts or Flagger with the Istio sidecar injection options above.

## VPA (Vertical Pod Autoscaler)

The chart does not install VPA. For right-sizing recommendations after go-live:

```bash
kubectl apply -f https://github.com/kubernetes/autoscaler/releases/latest/download/vertical-pod-autoscaler.yaml
# Create a VPA in "Off" or "Recommendation" mode targeting the pucora Deployment
```

## Upgrade, test, rollback

```bash
helm upgrade my-gateway ./deploy/helm/pucora -f my-values.yaml
helm test my-gateway
helm rollback my-gateway
helm uninstall my-gateway
```

## Common values

| Parameter | Default | Description |
|-----------|---------|-------------|
| `replicaCount` | `2` | Gateway replicas |
| `config.mode` | `configmap` | `configmap`, `secret`, or `image` |
| `usageReporting.disable` | `true` | Set `USAGE_DISABLE=1` |
| `probes.startup.enabled` | `false` | Startup probe |
| `lifecycle.preStop.enabled` | `true` | Graceful drain sleep |
| `service.nlb.enabled` | `false` | NLB LoadBalancer |
| `sidecarInjection.enabled` | `false` | Mesh injection annotations |
| `networkPolicy.enabled` | `false` | Ingress NetworkPolicy |
| `keda.enabled` | `false` | KEDA ScaledObject |
| `tests.configCheck` | `true` | `helm test` runs `pucora check` |

See [`values.yaml`](values.yaml) and [`values.schema.json`](values.schema.json) for the full list.

## Uninstall

```bash
helm uninstall my-gateway
```
