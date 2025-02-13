{{- define "common.annotations" -}}
{{- $annotations := dict -}}
{{- with .Values.podAnnotations }}
  {{- $annotations = merge $annotations . -}}
{{- end }}
{{- if .Values.tdxEnabled }}
  {{- with .Values.commonlibrary.tdx }}
    {{- $annotations = merge $annotations .common.annotations -}}
  {{- end }}
{{- end }}
{{- if gt (len $annotations) 0 -}}
annotations:
{{- toYaml $annotations | indent 2 }}
{{- end }}
{{- end }}

{{- define "common.runtimeClassName" -}}
{{- if .Values.tdxEnabled }}
  {{- with .Values.commonlibrary.tdx }}
runtimeClassName: {{ .common.runtimeClassName }}
  {{- end }}
{{- end }}
{{- end }}