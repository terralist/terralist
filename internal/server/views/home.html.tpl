{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <p>Welcome back, {{ .Values.User.Name }}!</p>
  <p>Your email: {{ .Values.User.Email }}</p>

  <a href="/logout">Logout</a>
</div>

{{ end }}

{{ define "script" }}
{{ end }}