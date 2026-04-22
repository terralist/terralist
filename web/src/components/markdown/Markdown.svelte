<script lang="ts">
  import { onDestroy, tick } from 'svelte';
  import { get } from 'svelte/store';
  import { location, querystring } from 'svelte-spa-router';

  import MarkdownCode from '@/components/MarkdownCode.svelte';
  import {
    CODE_PLACEHOLDER_CLASS,
    decodeBase64Utf8,
    MD_ANCHOR_QUERY,
    render
  } from '@/lib/markdown/render';

  export let source: string;

  let container: HTMLElement | null = null;
  let mounted: Array<{ $destroy: () => void }> = [];
  /** Skip tearing down Shiki/Mermaid when only `?md-anchor=` changed. */
  let lastMountedHtml: string | null = null;
  let highlightedForAnchor: Element | null = null;
  /** Cancels in-flight deep-link scroll when the component tears down or deps change. */
  let anchorScrollGeneration = 0;
  const anchorFollowUpTimers: ReturnType<typeof setTimeout>[] = [];

  const clearAnchorFollowUpTimers = (): void => {
    for (const t of anchorFollowUpTimers) {
      clearTimeout(t);
    }
    anchorFollowUpTimers.length = 0;
  };

  const clearMdAnchorHighlight = (): void => {
    highlightedForAnchor?.classList.remove('md-anchor-target');
    highlightedForAnchor = null;
  };

  const scrollElToView = (el: HTMLElement): void => {
    el.scrollIntoView({ behavior: 'auto', block: 'start' });
  };

  /**
   * After a cold open, headings exist in `{@html}` but the target may not be
   * in the DOM yet (async route), or layout shifts once Shiki/Mermaid paint.
   * Retry until the id appears, then nudge scroll a few times as layout settles.
   */
  const scrollToMdAnchorFromQuery = async (): Promise<void> => {
    const generation = ++anchorScrollGeneration;
    clearAnchorFollowUpTimers();
    clearMdAnchorHighlight();

    let id: string | null = null;
    try {
      id = new URLSearchParams(get(querystring) ?? '').get(MD_ANCHOR_QUERY);
      if (id) id = decodeURIComponent(id);
    } catch {
      id = null;
    }
    if (!id) return;

    const maxWaitMs = 4000;
    const stepMs = 32;
    const deadline = Date.now() + maxWaitMs;

    let el: HTMLElement | null = null;
    while (Date.now() < deadline && generation === anchorScrollGeneration) {
      await tick();
      el = document.getElementById(id);
      if (el) break;
      await new Promise<void>(r => setTimeout(r, stepMs));
    }

    if (generation !== anchorScrollGeneration || !el) return;

    const anchorId = id;
    highlightedForAnchor = el;
    el.classList.add('md-anchor-target');

    const nudgeScroll = (): void => {
      if (generation !== anchorScrollGeneration) return;
      const node = document.getElementById(anchorId);
      if (!node) return;
      scrollElToView(node);
    };

    nudgeScroll();
    requestAnimationFrame(nudgeScroll);
    requestAnimationFrame(() => requestAnimationFrame(nudgeScroll));
    for (const delay of [120, 320, 720]) {
      anchorFollowUpTimers.push(setTimeout(nudgeScroll, delay));
    }
  };

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

  $: html = render(source ?? '', { routePath: $location ?? '' });

  // Replace each `<div class="md-code-placeholder">` emitted by the marked
  // renderer with a `MarkdownCode` Svelte component instance, so Shiki and
  // Mermaid run for every code block. We tear down previously mounted
  // components on each render to avoid leaks across source changes (theme
  // toggles, version switches, etc.).
  const mountCodeBlocks = async (): Promise<void> => {
    await tick();
    if (!container) return;

    const htmlChanged = lastMountedHtml !== html;
    if (htmlChanged) {
      lastMountedHtml = html;
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
    }

    await tick();
    void scrollToMdAnchorFromQuery();
  };

  $: {
    void html;
    void $querystring;
    void mountCodeBlocks();
  }

  onDestroy(() => {
    anchorScrollGeneration += 1;
    clearAnchorFollowUpTimers();
    clearMdAnchorHighlight();
    for (const instance of mounted) {
      instance.$destroy();
    }
    mounted = [];
    lastMountedHtml = null;
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
