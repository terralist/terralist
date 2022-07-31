{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <h2>Sign In with {{ .Values.Provider }}</h2>
  <p>This resource is protected and requires authentication using your {{ .Values.Provider }} account.</p>
  {{ if (empty .Values.Error | not) }}
  <div class="error">
    <p>{{ .Values.ErrorDescription }}</p>
  </div>
  {{ end }}
  <form method="get" action={{ .Values.AuthorizationEndpoint }}>
    <input type="hidden" name="redirect_uri" value={{ .Values.HostURL }} />
    <button type="submit">Continue</button>
  </form>
</div>

{{ end }}

{{ define "script" }}
{{ end }}