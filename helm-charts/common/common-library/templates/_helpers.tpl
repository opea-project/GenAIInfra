{{- define "common.annotations" -}}
{{- $annotations := dict -}}
{{- with .Values.podAnnotations }}
  {{- $annotations = merge $annotations . -}}
{{- end }}
{{- with .Values.tdx }}
  {{- $annotations = merge $annotations .tdx.common.annotations -}}
{{- end }}
{{- if gt (len $annotations) 0 -}}
annotations:
{{- toYaml $annotations | nindent 2 }}
{{- end }}
{{- end }}

{{- define "common.runtimeClassName" -}}
{{- with .Values.tdx }}
runtimeClassName: {{ .tdx.common.runtimeClassName }}
{{- end }}
{{- end }}
