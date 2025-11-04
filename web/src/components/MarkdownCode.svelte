<script lang="ts">
  import { onMount } from 'svelte';
  import hljs from 'highlight.js';

  export let lang: string | undefined;
  export let text: string | undefined;

  let codeEl: HTMLElement | null = null;

  const applyHighlight = () => {
    if (!codeEl) return;
    // Reset content and classes before highlighting
    codeEl.className = 'hljs' + (lang ? ` language-${lang}` : '');
    codeEl.textContent = text ?? '';
    try {
      // Use explicit language if provided; otherwise, auto-detect
      if (lang && hljs.getLanguage(lang)) {
        hljs.highlightElement(codeEl);
      } else {
        hljs.highlightElement(codeEl);
      }
    } catch {}
  };

  onMount(applyHighlight);
  $: lang, text, applyHighlight();
</script>

<pre><code bind:this={codeEl}></code></pre>
