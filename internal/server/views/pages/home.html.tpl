{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <p>Welcome back, {{ $.Values.User.Name }}!</p>
  <p>Your email: {{ $.Values.User.Email }}</p>

  <a href="{{ $.Values.Endpoints.Logout }}">Logout</a>

  <div class="authorities">
    <h3 class="title">
      Authorities
      |
      <a href="{{ $.Values.Endpoints.CreateAuthority }}">Create</a>
    </h3>

    {{ if eq (len .Values.Authorities) 0 }}
    <p>No authorities found for this account.</p>
    {{ else }}
    <ol style="list-style-position: outside;">
      {{ range $i, $authority := .Values.Authorities }}
      <li>
        
        <h4>
          {{ $authority.Name }}
          |
          <a href="{{ $authority.PolicyURL }}">Policy</a>
          |
          <a href="{{ $.Values.Endpoints.RemoveAuthority | replace ":id" (toString $authority.ID) }}">Remove</a>
        </h4>

        <p>
          Keys
          |
          <a href="{{ $.Values.Endpoints.CreateKey | replace ":id" (toString $authority.ID) }}">Add</a>
        </p>
        {{ if eq (len $authority.Keys) 0 }}
        <p>No key found for this authority.</p>
        {{ else }}
        <ul>
          {{ range $j, $key := $authority.Keys }}
            <li>
              ID: {{ $key.KeyId }} 
              | 
              <a href="{{ $.Values.Endpoints.RemoveKey | replace ":id" (toString $authority.ID) | replace ":kid" (toString $key.ID) }}">
                Remove
              </a>
            </li>
          {{ end }}
        </ul>
        {{ end }}

        <p>
          API Keys
          |
          <a href="{{ $.Values.Endpoints.CreateApiKey | replace ":id" (toString $authority.ID) }}">Add</a>
        </p>
        {{ if eq (len $authority.ApiKeys) 0 }}
        <p>No API key found for this authority.</p>
        {{ else }}
        <ul>
          {{ range $j, $apiKey := $authority.ApiKeys }}
            <li>
              ID: {{ $apiKey.ID }} 
              | 
              <a href="{{ $.Values.Endpoints.RemoveApiKey | replace ":id" (toString $authority.ID) | replace ":kid" (toString $apiKey.ID) }}">
                Remove
              </a>
            </li>
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