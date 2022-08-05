<!DOCTYPE html>
<html>
  <head>
    <title>Terralist | {{ .Values.Title | default "Home" }}</title>

    <style>
      {{ block "style" .}}
      {{ end }}

      .error .headline {
        color: red;
      }
      
      .error .description {
        color: lightcoral;
      }
    </style>
  </head>

  <body>
    {{ if not (empty .Values.Error) }}
    <div class="error">
      <p class="headline">{{ .Values.Error.Name | default "An unknown error occurred" }}</p>
      {{ if (empty .Values.Error.Description | not) }}
      <p class="description">{{ .Values.Error.Description }}</p>
      {{ end }}
    </div>
    {{ end }}

    <div class="main">
      {{ block "content" . }}
        <!-- Fallback -->
        <h2>404 not found</h2>
      {{ end }}
    </div>

    <script type="text/javascript">
      {{ block "script" . }}
      {{ end }}
    </script>
  </body>
</html>