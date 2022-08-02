{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <p>Welcome back, {{ .Values.User.Name }}!</p>
  <p>Your email: {{ .Values.User.Email }}</p>

  <a href="/logout">Logout</a>

  <div class="authorities">
    <h3 class="title">
      Authorities
      |
      <a href="/authority/create">Create</a>
    </h3>

    {{ if eq (len .Values.Authorities) 0 }}
    <p>No authorities found for this account.</p>
    {{ else }}
    <ol style="list-style-position: outside;">
      {{ range .Values.Authorities }}
      <li>
        
        <h4>
          {{ .Name }}
          |
          <a href="{{ .PolicyURL }}">Policy</a>
          |
          <a href="/authority/delete/{{ .ID }}">Remove</a>
        </h4>

        <p>
          Keys
          |
          <a href="/authority/{{ .ID }}/keys/add">Add</a>
        </p>
        <ul>
          {{ range .Keys }}
            <li>ID: {{ .KeyId }} | <a href="/authority/keys/{{ .ID }}/remove">Remove</a></li>
          {{ end }}
        </ul>

        <p>
          API Keys
          |
          <a href="/authority/{{ .ID }}/apikeys/add">Add</a>
        </p>
        {{ if eq (len .ApiKeys) 0 }}
        <p>No API keys found for this authority.</p>
        {{ else }}
        <ul>
          {{ range .ApiKeys }}
            <li>ID: {{ .ID }} | <a href="/authority/apikeys/{{ .ID }}/remove">Remove</a></li>
          {{ end }}
        </ul>
        {{ end }}

      </li>
      {{ end }}
    </ol>
    {{ end }}
  </div>
</div>

{{ end }}

{{ define "script" }}
{{ end }}