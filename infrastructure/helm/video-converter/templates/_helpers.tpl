{{/*
Expand the name of the chart.
*/}}
{{- define "video-converter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "video-converter.fullname" -}}
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
{{- define "video-converter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "video-converter.labels" -}}
helm.sh/chart: {{ include "video-converter.chart" . }}
{{ include "video-converter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "video-converter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "video-converter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "video-converter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "video-converter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Gateway service labels
*/}}
{{- define "video-converter.gateway.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: gateway
{{- end }}

{{/*
Gateway selector labels
*/}}
{{- define "video-converter.gateway.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: gateway
{{- end }}

{{/*
Auth service labels
*/}}
{{- define "video-converter.auth.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: auth
{{- end }}

{{/*
Auth selector labels
*/}}
{{- define "video-converter.auth.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: auth
{{- end }}

{{/*
Converter service labels
*/}}
{{- define "video-converter.converter.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: converter
{{- end }}

{{/*
Converter selector labels
*/}}
{{- define "video-converter.converter.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: converter
{{- end }}

{{/*
Analytics service labels
*/}}
{{- define "video-converter.analytics.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: analytics
{{- end }}

{{/*
Analytics selector labels
*/}}
{{- define "video-converter.analytics.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: analytics
{{- end }}

{{/*
Realtime service labels
*/}}
{{- define "video-converter.realtime.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: realtime
{{- end }}

{{/*
Realtime selector labels
*/}}
{{- define "video-converter.realtime.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: realtime
{{- end }}

{{/*
Frontend service labels
*/}}
{{- define "video-converter.frontend.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "video-converter.frontend.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Notification service labels
*/}}
{{- define "video-converter.notification.labels" -}}
{{ include "video-converter.labels" . }}
app.kubernetes.io/component: notification
{{- end }}

{{/*
Notification selector labels
*/}}
{{- define "video-converter.notification.selectorLabels" -}}
{{ include "video-converter.selectorLabels" . }}
app.kubernetes.io/component: notification
{{- end }}

{{/*
Image name helper
*/}}
{{- define "video-converter.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .repository .tag }}
{{- else }}
{{- printf "%s:%s" .repository .tag }}
{{- end }}
{{- end }}