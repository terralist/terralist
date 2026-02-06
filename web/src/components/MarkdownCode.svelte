<script lang="ts">
  /* eslint-disable svelte/no-dom-manipulating */
  import { onDestroy, onMount } from 'svelte';
  import context, { type Theme } from '@/context';

  type ShikiHighlighter = {
    codeToHtml: (
      code: string,
      options: { lang?: string; theme?: string; transformers?: unknown[] }
    ) => Promise<string>;
  };

  type ShikiModule = {
    createHighlighter: (options: {
      themes?: readonly string[];
      langs?: readonly string[];
    }) => Promise<ShikiHighlighter>;
  };

  type MermaidApi = {
    initialize: (config: {
      startOnLoad?: boolean;
      securityLevel?: 'strict' | 'loose' | 'antiscript';
      theme?: 'default' | 'dark' | 'forest' | 'neutral';
    }) => void;
    render: (id: string, text: string) => Promise<{ svg: string }>;
  };

  export let lang: string | undefined;
  export let text: string | undefined;
  export let shiki: ShikiHighlighter | null = null;
  export let mermaid: MermaidApi | null = null;

  let codeEl: HTMLElement | null = null;
  let mermaidEl: HTMLElement | null = null;
  let currentTheme: Theme = 'light';
  const unsubscribe = context.theme.subscribe(t => (currentTheme = t));

  let shikiHighlighter: ShikiHighlighter | null = null;
  let mermaidFallback: MermaidApi | null = null;
  let shikiLoading: Promise<void> | null = null;
  let mermaidLoading: Promise<void> | null = null;

  const ensureShiki = async (): Promise<ShikiHighlighter | null> => {
    if (shiki) return shiki;
    if (shikiHighlighter) return shikiHighlighter;
    if (!shikiLoading) {
      shikiLoading = import('https://cdn.jsdelivr.net/npm/shiki@3.15.0/+esm')
        .then(async ({ createHighlighter }: ShikiModule) => {
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
          console.error(
            'Failed to load Shikijs from CDN in MarkdownCode:',
            err
          );
        })
        .finally(() => {
          shikiLoading = null;
        });
    }

    await shikiLoading;
    return shiki || shikiHighlighter;
  };

  const ensureMermaid = async (): Promise<MermaidApi | null> => {
    if (mermaid) return mermaid;
    if (mermaidFallback) return mermaidFallback;
    if (!mermaidLoading) {
      mermaidLoading =
        import('https://cdn.jsdelivr.net/npm/mermaid@10.9.0/+esm')
          .then(module => {
            mermaidFallback = module.default as MermaidApi;
          })
          .catch(err => {
            console.error(
              'Failed to load mermaid from CDN in MarkdownCode:',
              err
            );
          })
          .finally(() => {
            mermaidLoading = null;
          });
    }

    await mermaidLoading;
    return mermaid || mermaidFallback;
  };

  const applyHighlight = async () => {
    if (!codeEl || !text) return;
    codeEl.textContent = text ?? '';
    try {
      const shikiInstance = shiki || (await ensureShiki());
      if (shikiInstance) {
        const theme = currentTheme === 'dark' ? 'github-dark' : 'github-light';
        const html = await shikiInstance.codeToHtml(text, {
          lang: lang || 'text',
          theme,
          transformers: []
        });
        codeEl.innerHTML = html;
      }
    } catch (err) {
      console.error('Failed to highlight code block:', err);
    }
  };

  const renderMermaid = async () => {
    if (!mermaidEl || !text) return;
    try {
      const mermaidInstance = mermaid || (await ensureMermaid());
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
    } catch (err) {
      console.error('Failed to render mermaid diagram:', err);
    }
  };

  onMount(async () => {
    if (lang === 'mermaid') {
      await renderMermaid();
    } else {
      await applyHighlight();
    }
  });

  // Trigger re-render when content or theme changes
  $: {
    if (lang === 'mermaid') {
      renderMermaid();
    } else if (text) {
      applyHighlight();
    }
  }

  onDestroy(() => unsubscribe());
</script>

{#if lang === 'mermaid'}
  <div class="mermaid" bind:this={mermaidEl}></div>
{:else}
  <pre><code bind:this={codeEl}></code></pre>
{/if}
