{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <h2>Create Authority</h2>
  
  <br />

  <form method="post" action="#">
    <div style="display: flex; flex-direction: row;">
      <label for="name">Name:</p>
      <input type="text" name="name" id="name" />
    </div>
    <br/>

    <div style="display: flex; flex-direction: row;">
      <label for="policy_url">Policy URL:</p>
      <input type="text" name="policy_url" id="policy_url" />
    </div>
    <br />
    
    <button type="submit">Create</button>
  </form>
</div>

{{ end }}

{{ define "script" }}
{{ end }}