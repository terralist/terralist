<script lang="ts">
  import { onDestroy, tick } from 'svelte';

  import MarkdownCode from '@/components/MarkdownCode.svelte';
  import {
    CODE_PLACEHOLDER_CLASS,
    decodeBase64Utf8,
    render
  } from '@/lib/markdown/render';

  export let source: string;

  let container: HTMLElement | null = null;
  let mounted: Array<{ $destroy: () => void }> = [];

  /**
   * svelte-spa-router keeps the app path in `location.hash` as `#/route`.
   * Plain fragment links (`href="#footnote-def-1"`, heading `#tables`, …)
   * would replace that entire hash and kick the user off the current page.
   * Intercept those and scroll inside the document instead.
   */
  const handleInDocHashClick = (e: MouseEvent): void => {
    if (e.defaultPrevented || e.button !== 0) return;
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

    const pathTarget = e.composedPath?.()[0] ?? e.target;
    if (!(pathTarget instanceof Element)) return;
    const anchor = pathTarget.closest('a');
    if (!anchor || !(anchor instanceof HTMLAnchorElement)) return;

    const raw = anchor.getAttribute('href');
    if (raw == null) return;
    const href = raw.trim();
    if (!href.startsWith('#') || href.startsWith('#/')) return;
    if (href === '#') return;

    let id: string;
    try {
      id = decodeURIComponent(href.slice(1));
    } catch {
      return;
    }
    if (!id) return;

    const el = document.getElementById(id);
    if (!el) return;

    e.preventDefault();
    el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  $: html = render(source ?? '');

  // Replace each `<div class="md-code-placeholder">` emitted by the marked
  // renderer with a `MarkdownCode` Svelte component instance, so Shiki and
  // Mermaid run for every code block. We tear down previously mounted
  // components on each render to avoid leaks across source changes (theme
  // toggles, version switches, etc.).
  const mountCodeBlocks = async (): Promise<void> => {
    await tick();
    if (!container) return;

    for (const instance of mounted) {
      instance.$destroy();
    }
    mounted = [];

    const placeholders = container.querySelectorAll<HTMLElement>(
      `.${CODE_PLACEHOLDER_CLASS}`
    );
    placeholders.forEach(placeholder => {
      const lang = placeholder.dataset.mdLang || undefined;
      const text = decodeBase64Utf8(placeholder.dataset.mdCode ?? '');

      const target = document.createElement('div');
      target.className = 'md-code-mount';
      placeholder.replaceWith(target);

      mounted.push(
        new MarkdownCode({
          target,
          props: { lang, text }
        })
      );
    });
  };

  $: {
    void html;
    void mountCodeBlocks();
  }

  onDestroy(() => {
    for (const instance of mounted) {
      instance.$destroy();
    }
    mounted = [];
  });
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div
  class="markdown-body"
  bind:this={container}
  on:click|capture={handleInDocHashClick}>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
  {@html html}
</div>
