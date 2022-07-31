{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <h2>Error</h2>
  <p>{{ .Values.Error | default "An unknown error occurred" }}</p>
  {{ if (empty .Values.Description | not) }}
  <p>{{ .Values.Description }}</p>
  {{ end }}
</div>

{{ end }}

{{ define "script" }}
{{ end }}