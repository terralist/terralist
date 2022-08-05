{{ define "style" }}
{{ end }}

{{ define "content" }}

<div class="container">
  <h2>Add Key for Authority</h2>
  
  <br />

  <form method="post" action="#">
    <div style="display: flex; flex-direction: row;">
      <label for="key_id">Key ID:</p>
      <input type="text" name="key_id" id="key_id" />
    </div>
    <br/>

    <div style="display: flex; flex-direction: row;">
      <label for="ascii_armor">ASCII Armor:</p>
      <textarea type="text" name="ascii_armor" id="ascii_armor"></textarea>
    </div>
    <br />

    <div style="display: flex; flex-direction: row;">
      <label for="trust_signature">Trust Signature:</p>
      <input type="text" name="trust_signature" id="trust_signature" />
    </div>
    <br />
    
    <button type="submit">Create</button>
  </form>
</div>

{{ end }}

{{ define "script" }}
{{ end }}