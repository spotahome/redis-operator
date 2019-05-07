{{/* Build the Spotahome standard labels */}}
{{- define "common-labels" -}}
app.kubernetes.io/name: {{ .Chart.Name | quote }}
{{- end }}

{{- define "helm-labels" -}}
{{ include "common-labels" . }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- end }}

{{/* Build wide-used variables the application */}}
{{ define "name" -}}
{{ printf "%s-%s" .Release.Name .Chart.Name }}
{{- end }}

{{ define "image" -}}
{{ printf "%s:%s" .Values.image .Values.tag }}
{{- end }}
