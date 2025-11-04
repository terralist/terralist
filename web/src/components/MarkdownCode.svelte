<script lang="ts">
  /* eslint-disable svelte/no-dom-manipulating, no-empty, @typescript-eslint/no-explicit-any */
  import { onDestroy, onMount } from 'svelte';
  import hljs from 'highlight.js';
  import mermaid from 'mermaid';
  import context, { type Theme } from '@/context';
  import hljsTerraform, {
    definer as hljsTerraformDefiner
  } from '@/lib/hljs-terraform';

  export let lang: string | undefined;
  export let text: string | undefined;

  let codeEl: HTMLElement | null = null;
  let mermaidEl: HTMLElement | null = null;
  let currentTheme: Theme = 'light';
  const unsubscribe = context.theme.subscribe(t => (currentTheme = t));

  // Register Terraform/HCL grammar if available (vendored)
  try {
    // Register terraform
    if (hljsTerraform && typeof hljsTerraform === 'function') {
      hljsTerraform(hljs);
    }
    // Register aliases using the definer
    const def = hljsTerraformDefiner as any;
    if (def) {
      if (!hljs.getLanguage('hcl')) hljs.registerLanguage('hcl', def);
      if (!hljs.getLanguage('tf')) hljs.registerLanguage('tf', def);
    }
  } catch {}

  const applyHighlight = () => {
    if (!codeEl) return;
    codeEl.className = 'hljs' + (lang ? ` language-${lang}` : '');
    codeEl.textContent = text ?? '';
    try {
      hljs.highlightElement(codeEl);
    } catch {}
  };

  const renderMermaid = async () => {
    if (!mermaidEl || !text) return;
    try {
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: 'strict',
        theme: currentTheme === 'dark' ? 'dark' : 'default'
      });

      const id = `mmd-${Math.random().toString(36).slice(2)}`;
      const { svg } = await mermaid.render(id, text);
      mermaidEl.innerHTML = svg;
    } catch {}
  };

  onMount(() => {
    if (lang === 'mermaid') {
      renderMermaid();
    } else {
      applyHighlight();
    }
  });

  $: if (lang === 'mermaid') {
    renderMermaid();
  } else {
    applyHighlight();
  }

  onDestroy(() => unsubscribe());
</script>

{#if lang === 'mermaid'}
  <div class="mermaid" bind:this={mermaidEl}></div>
{:else}
  <pre><code bind:this={codeEl}></code></pre>
{/if}
