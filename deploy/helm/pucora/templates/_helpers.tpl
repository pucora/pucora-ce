{{/*
Expand the name of the chart.
*/}}
{{- define "pucora.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "pucora.fullname" -}}
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
{{- define "pucora.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "pucora.labels" -}}
helm.sh/chart: {{ include "pucora.chart" . }}
{{ include "pucora.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "pucora.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pucora.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "pucora.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "pucora.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
ConfigMap name for pucora.json
*/}}
{{- define "pucora.configMapName" -}}
{{- if .Values.config.existingConfigMap }}
{{- .Values.config.existingConfigMap }}
{{- else }}
{{- include "pucora.fullname" . }}
{{- end }}
{{- end }}

{{/*
Secret name for pucora.json
*/}}
{{- define "pucora.configSecretName" -}}
{{- if .Values.config.existingSecret }}
{{- .Values.config.existingSecret }}
{{- else }}
{{- include "pucora.fullname" . }}
{{- end }}
{{- end }}

{{/*
Whether configuration is mounted from a ConfigMap.
*/}}
{{- define "pucora.useConfigMap" -}}
{{- if and (eq .Values.config.mode "configmap") (or .Values.config.existingConfigMap .Values.config.pucoraJson) }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Whether configuration is mounted from a Secret.
*/}}
{{- define "pucora.useConfigSecret" -}}
{{- if and (eq .Values.config.mode "secret") (or .Values.config.existingSecret .Values.config.pucoraJson) }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Whether configuration is mounted from ConfigMap or Secret.
*/}}
{{- define "pucora.useConfigVolume" -}}
{{- if or (eq (include "pucora.useConfigMap" .) "true") (eq (include "pucora.useConfigSecret" .) "true") }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Chart-managed config (not external ConfigMap/Secret).
*/}}
{{- define "pucora.chartManagedConfig" -}}
{{- if and (or (eq .Values.config.mode "configmap") (eq .Values.config.mode "secret")) (not .Values.config.existingConfigMap) (not .Values.config.existingSecret) }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Metrics service name
*/}}
{{- define "pucora.metricsServiceName" -}}
{{- printf "%s-metrics" (include "pucora.fullname" .) }}
{{- end }}

{{/*
Service type (NLB mode forces LoadBalancer).
*/}}
{{- define "pucora.serviceType" -}}
{{- if .Values.service.nlb.enabled -}}
LoadBalancer
{{- else -}}
{{ .Values.service.type }}
{{- end -}}
{{- end }}

{{/*
Whether the main Service is a LoadBalancer (NLB or explicit type).
*/}}
{{- define "pucora.isLoadBalancerService" -}}
{{- if or .Values.service.nlb.enabled (eq .Values.service.type "LoadBalancer") -}}
true
{{- else -}}
false
{{- end -}}
{{- end }}

{{/*
Merged Service annotations (base + NLB-specific).
*/}}
{{- define "pucora.serviceAnnotations" -}}
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
{{- define "pucora.sidecarInjectionAnnotations" -}}
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
{{- define "pucora.gitopsAnnotations" -}}
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
{{- define "pucora.certificateSecretName" -}}
{{- if .Values.certificate.secretName }}
{{- .Values.certificate.secretName }}
{{- else }}
{{- printf "%s-tls" (include "pucora.fullname" .) }}
{{- end }}
{{- end }}
