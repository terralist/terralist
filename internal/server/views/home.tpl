{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <p>Welcome back, {{ .Values.User.Name }}!</p>
  <p>Your email: {{ .Values.User.Email }}</p>
</div>

{{ end }}

{{ define "script" }}
{{ end }}