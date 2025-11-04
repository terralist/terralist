<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import hljs from 'highlight.js';
  import mermaid from 'mermaid';
  import context, { type Theme } from '@/context';

  export let lang: string | undefined;
  export let text: string | undefined;

  let codeEl: HTMLElement | null = null;
  let mermaidEl: HTMLElement | null = null;
  let currentTheme: Theme = 'light';
  const unsubscribe = context.theme.subscribe(t => (currentTheme = t));

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
