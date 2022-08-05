{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <h2>Sign In with {{ .Values.Provider }}</h2>
  <p>This resource is protected and requires authentication using your {{ .Values.Provider }} account.</p>
  <form method="get" action={{ .Values.Endpoints.Authorization }}>
    <input type="hidden" name="redirect_uri" value={{ .Values.HostURL }} />
    <button type="submit">Continue</button>
  </form>
</div>

{{ end }}

{{ define "script" }}
{{ end }}