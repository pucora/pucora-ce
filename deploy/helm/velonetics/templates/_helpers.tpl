{{/*
Expand the name of the chart.
*/}}
{{- define "velonetics.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "velonetics.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "velonetics.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "velonetics.labels" -}}
helm.sh/chart: {{ include "velonetics.chart" . }}
{{ include "velonetics.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "velonetics.selectorLabels" -}}
app.kubernetes.io/name: {{ include "velonetics.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "velonetics.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "velonetics.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
ConfigMap name for velonetics.json
*/}}
{{- define "velonetics.configMapName" -}}
{{- if .Values.config.existingConfigMap }}
{{- .Values.config.existingConfigMap }}
{{- else }}
{{- include "velonetics.fullname" . }}
{{- end }}
{{- end }}

{{/*
Secret name for velonetics.json
*/}}
{{- define "velonetics.configSecretName" -}}
{{- if .Values.config.existingSecret }}
{{- .Values.config.existingSecret }}
{{- else }}
{{- include "velonetics.fullname" . }}
{{- end }}
{{- end }}

{{/*
Whether configuration is mounted from a ConfigMap.
*/}}
{{- define "velonetics.useConfigMap" -}}
{{- if and (eq .Values.config.mode "configmap") (or .Values.config.existingConfigMap .Values.config.veloneticsJson) }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Whether configuration is mounted from a Secret.
*/}}
{{- define "velonetics.useConfigSecret" -}}
{{- if and (eq .Values.config.mode "secret") (or .Values.config.existingSecret .Values.config.veloneticsJson) }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Whether configuration is mounted from ConfigMap or Secret.
*/}}
{{- define "velonetics.useConfigVolume" -}}
{{- if or (eq (include "velonetics.useConfigMap" .) "true") (eq (include "velonetics.useConfigSecret" .) "true") }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Chart-managed config (not external ConfigMap/Secret).
*/}}
{{- define "velonetics.chartManagedConfig" -}}
{{- if and (or (eq .Values.config.mode "configmap") (eq .Values.config.mode "secret")) (not .Values.config.existingConfigMap) (not .Values.config.existingSecret) }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Metrics service name
*/}}
{{- define "velonetics.metricsServiceName" -}}
{{- printf "%s-metrics" (include "velonetics.fullname" .) }}
{{- end }}

{{/*
Service type (NLB mode forces LoadBalancer).
*/}}
{{- define "velonetics.serviceType" -}}
{{- if .Values.service.nlb.enabled -}}
LoadBalancer
{{- else -}}
{{ .Values.service.type }}
{{- end -}}
{{- end }}

{{/*
Whether the main Service is a LoadBalancer (NLB or explicit type).
*/}}
{{- define "velonetics.isLoadBalancerService" -}}
{{- if or .Values.service.nlb.enabled (eq .Values.service.type "LoadBalancer") -}}
true
{{- else -}}
false
{{- end -}}
{{- end }}

{{/*
Merged Service annotations (base + NLB-specific).
*/}}
{{- define "velonetics.serviceAnnotations" -}}
{{- $annotations := .Values.service.annotations | deepCopy -}}
{{- if .Values.service.nlb.enabled -}}
{{- $annotations = merge $annotations (.Values.service.nlb.annotations | default dict) -}}
{{- end -}}
{{- if $annotations -}}
{{- toYaml $annotations -}}
{{- end -}}
{{- end }}

{{/*
Sidecar injection annotations for the pod template.
*/}}
{{- define "velonetics.sidecarInjectionAnnotations" -}}
{{- if .Values.sidecarInjection.enabled -}}
{{- if .Values.sidecarInjection.istio.enabled -}}
sidecar.istio.io/inject: {{ .Values.sidecarInjection.istio.inject | ternary "true" "false" | quote }}
{{- if .Values.sidecarInjection.istio.holdApplicationUntilProxyStarts }}
proxy.istio.io/config: '{ "holdApplicationUntilProxyStarts": true }'
{{- end }}
{{- end -}}
{{- if .Values.sidecarInjection.linkerd.enabled }}
linkerd.io/inject: {{ .Values.sidecarInjection.linkerd.inject | quote }}
{{- end -}}
{{- with .Values.sidecarInjection.annotations }}
{{- toYaml . }}
{{- end -}}
{{- end -}}
{{- end }}

{{/*
GitOps annotations for resources.
*/}}
{{- define "velonetics.gitopsAnnotations" -}}
{{- if .Values.gitops.argocd.enabled }}
{{- if .Values.gitops.argocd.syncWave }}
argocd.argoproj.io/sync-wave: {{ .Values.gitops.argocd.syncWave | quote }}
{{- end }}
{{- end }}
{{- if .Values.gitops.flux.enabled }}
{{- if .Values.gitops.flux.reconcile }}
reconcile.fluxcd.io/requestedAt: {{ .Values.gitops.flux.reconcile | quote }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Certificate secret name for Ingress TLS.
*/}}
{{- define "velonetics.certificateSecretName" -}}
{{- if .Values.certificate.secretName }}
{{- .Values.certificate.secretName }}
{{- else }}
{{- printf "%s-tls" (include "velonetics.fullname" .) }}
{{- end }}
{{- end }}
