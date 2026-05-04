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
  let copied = false;
  let copyResetTimer: ReturnType<typeof setTimeout> | null = null;

  const handleCopy = async (): Promise<void> => {
    const value = text ?? '';
    if (!value) return;
    try {
      if (
        typeof navigator !== 'undefined' &&
        navigator.clipboard &&
        typeof navigator.clipboard.writeText === 'function'
      ) {
        await navigator.clipboard.writeText(value);
      } else {
        // Fallback for environments without async clipboard (older Safari,
        // insecure contexts). Uses a hidden textarea + execCommand.
        const ta = document.createElement('textarea');
        ta.value = value;
        ta.setAttribute('readonly', '');
        ta.style.position = 'absolute';
        ta.style.left = '-9999px';
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
      }
      copied = true;
      if (copyResetTimer) clearTimeout(copyResetTimer);
      copyResetTimer = setTimeout(() => {
        copied = false;
      }, 1500);
    } catch (err) {
      console.error('Failed to copy code block to clipboard:', err);
    }
  };

  // Auto-subscribe to the theme store so any `$:` block or template
  // expression that references `$theme` is tracked by Svelte's reactivity.
  const theme = context.theme;

  // Monotonic token so a stale async highlight (spawned before a theme
  // toggle) cannot overwrite a newer one. Without this, switching
  // dark -> light can land on dark output when the older Shiki call
  // resolves last.
  let renderToken = 0;

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

  const applyHighlight = async (activeTheme: Theme): Promise<void> => {
    if (!codeEl || !text) return;
    const token = ++renderToken;
    try {
      const shikiInstance = shiki || (await ensureShiki());
      if (token !== renderToken || !codeEl) return;
      if (!shikiInstance) return;
      const shikiTheme =
        activeTheme === 'dark' ? 'github-dark' : 'github-light';
      const html = await shikiInstance.codeToHtml(text, {
        lang: lang || 'text',
        theme: shikiTheme,
        transformers: []
      });
      if (token !== renderToken || !codeEl) return;
      codeEl.innerHTML = html;
    } catch (err) {
      console.error('Failed to highlight code block:', err);
    }
  };

  const renderMermaid = async (activeTheme: Theme): Promise<void> => {
    if (!mermaidEl || !text) return;
    const token = ++renderToken;
    try {
      const mermaidInstance = mermaid || (await ensureMermaid());
      if (token !== renderToken || !mermaidEl) return;
      if (!mermaidInstance) return;
      mermaidInstance.initialize({
        startOnLoad: false,
        securityLevel: 'strict',
        theme: activeTheme === 'dark' ? 'dark' : 'default'
      });

      const id = `mmd-${Math.random().toString(36).slice(2)}`;
      const { svg } = await mermaidInstance.render(id, text);
      if (token !== renderToken || !mermaidEl) return;
      mermaidEl.innerHTML = svg;
    } catch (err) {
      console.error('Failed to render mermaid diagram:', err);
    }
  };

  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  onDestroy(() => {
    if (copyResetTimer) clearTimeout(copyResetTimer);
  });

  // Re-render whenever the source, language, or theme changes. Guarded on
  // `mounted` so we don't race with `bind:this` before the DOM node exists.
  $: if (mounted) {
    if (lang === 'mermaid') {
      renderMermaid($theme);
    } else if (text) {
      applyHighlight($theme);
    }
  }
</script>

{#if lang === 'mermaid'}
  <div class="md-mermaid" bind:this={mermaidEl}></div>
{:else}
  <div class="md-code-block">
    <button
      class="md-copy-btn"
      class:is-copied={copied}
      type="button"
      aria-label={copied ? 'Copied' : 'Copy code'}
      title={copied ? 'Copied!' : 'Copy'}
      on:click={handleCopy}>
      {#if copied}
        <svg
          viewBox="0 0 16 16"
          width="16"
          height="16"
          aria-hidden="true"
          focusable="false">
          <path
            fill="currentColor"
            d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z" />
        </svg>
      {:else}
        <svg
          viewBox="0 0 16 16"
          width="16"
          height="16"
          aria-hidden="true"
          focusable="false">
          <path
            fill="currentColor"
            d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 0 1 0 1.5h-1.5a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 0 0 .25-.25v-1.5a.75.75 0 0 1 1.5 0v1.5A1.75 1.75 0 0 1 9.25 16h-7.5A1.75 1.75 0 0 1 0 14.25Z" />
          <path
            fill="currentColor"
            d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0 1 14.25 11h-7.5A1.75 1.75 0 0 1 5 9.25Zm1.75-.25a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 0 0 .25-.25v-7.5a.25.25 0 0 0-.25-.25Z" />
        </svg>
      {/if}
    </button>
    <div class="md-code-content" bind:this={codeEl}>
      <pre><code>{text ?? ''}</code></pre>
    </div>
  </div>
{/if}
