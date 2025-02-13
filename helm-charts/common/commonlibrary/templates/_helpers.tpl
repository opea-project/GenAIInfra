{{- define "common.annotations" -}}
{{- $annotations := dict -}}
{{- with .Values.podAnnotations }}
  {{- $annotations = merge $annotations . -}}
{{- end }}
{{- if .Values.tdxEnabled }}
    {{- $annotations = merge $annotations .Values.commonlibrary.tdx.annotations -}}
{{- end }}
{{- if gt (len $annotations) 0 -}}
annotations:
{{- toYaml $annotations | nindent 2 }}
{{- end }}
{{- end }}

{{- define "common.runtimeClassName" -}}
{{- if .Values.tdxEnabled }}
runtimeClassName: {{ .Values.commonlibrary.tdx.runtimeClassName }}
{{- end }}
{{- end }}