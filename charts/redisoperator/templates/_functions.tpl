{{/* Build the Spotahome standard labels */}}
{{- define "common-labels" -}}
app: {{ .Chart.Name | quote }}
team: {{ .Values.team | quote }}
{{- end }}

{{- define "helm-labels" -}}
{{ include "common-labels" . }}
chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
release: {{ .Release.Name | quote }}
heritage: {{ .Release.Service | quote }}
{{- end }}

{{/* Build wide-used variables the application */}}
{{ define "name" -}}
{{ printf "%s-%s" .Release.Name .Chart.Name }}
{{- end }}

{{ define "image" -}}
{{ printf "%s:%s" .Values.image .Values.tag }}
{{- end }}
