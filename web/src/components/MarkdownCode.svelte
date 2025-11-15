<script lang="ts">
  /* eslint-disable svelte/no-dom-manipulating, no-empty, @typescript-eslint/no-explicit-any */
  import { onDestroy, onMount } from 'svelte';
  import context, { type Theme } from '@/context';

  // Fallback imports for backward compatibility
  let shikiHighlighter: any = null;
  let mermaidFallback: any = null;

  export let lang: string | undefined;
  export let text: string | undefined;
  export let shiki: any = null;
  export let mermaid: any = null;

  let codeEl: HTMLElement | null = null;
  let mermaidEl: HTMLElement | null = null;
  let currentTheme: Theme = 'light';
  const unsubscribe = context.theme.subscribe(t => (currentTheme = t));

  // Load fallback libraries if not provided as props
  $: if (!shiki && !shikiHighlighter) {
    import('https://cdn.jsdelivr.net/npm/shiki@3.15.0/+esm')
      .then(async ({ createHighlighter }) => {
        shikiHighlighter = await createHighlighter({
          themes: ['github-light', 'github-dark'],
          langs: [
            'javascript',
            'typescript',
            'python',
            'bash',
            'json',
            'yaml',
            'markdown',
            'hcl',
            'terraform'
          ]
        });
      })
      .catch(err => {
        console.error('Failed to load Shikijs from CDN in MarkdownCode:', err);
      });
  }

  $: if (!mermaid && !mermaidFallback) {
    import('https://cdn.jsdelivr.net/npm/mermaid@10.9.0/+esm')
      .then(module => {
        mermaidFallback = module.default;
      })
      .catch(err => {
        console.error('Failed to load mermaid from CDN in MarkdownCode:', err);
      });
  }

  const applyHighlight = async () => {
    if (!codeEl) return;
    codeEl.textContent = text ?? '';
    try {
      const shikiInstance = shiki || shikiHighlighter;
      if (shikiInstance && text) {
        const theme = currentTheme === 'dark' ? 'github-dark' : 'github-light';
        const html = await shikiInstance.codeToHtml(text, {
          lang: lang || 'text',
          theme,
          transformers: []
        });
        codeEl.innerHTML = html;
      }
    } catch {}
  };

  const renderMermaid = async () => {
    if (!mermaidEl || !text) return;
    try {
      const mermaidInstance = mermaid || mermaidFallback;
      if (mermaidInstance) {
        mermaidInstance.initialize({
          startOnLoad: false,
          securityLevel: 'strict',
          theme: currentTheme === 'dark' ? 'dark' : 'default'
        });

        const id = `mmd-${Math.random().toString(36).slice(2)}`;
        const { svg } = await mermaidInstance.render(id, text);
        mermaidEl.innerHTML = svg;
      }
    } catch {}
  };

  onMount(async () => {
    if (lang === 'mermaid') {
      renderMermaid();
    } else {
      await applyHighlight();
    }
  });

  $: if (lang === 'mermaid') {
    renderMermaid();
  } else {
    // Trigger re-highlight when theme or content changes
    applyHighlight();
  }

  onDestroy(() => unsubscribe());
</script>

{#if lang === 'mermaid'}
  <div class="mermaid" bind:this={mermaidEl}></div>
{:else}
  <pre><code bind:this={codeEl}></code></pre>
{/if}
