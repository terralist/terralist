<script lang="ts">
  import { onDestroy } from 'svelte';
  import { emojify } from 'node-emoji';

  import context, { type Theme } from '@/context';
  import Navbar from '@/components/Navbar.svelte';
  import Markdown from '@/components/markdown/Markdown.svelte';
  import { markdownSample } from '@/lib/markdown/sample';

  import lightCssUrl from 'github-markdown-css/github-markdown-light.css?url';
  import darkCssUrl from 'github-markdown-css/github-markdown-dark.css?url';

  let currentTheme: Theme = 'light';
  const unsubscribeTheme = context.theme.subscribe(t => {
    currentTheme = t;
  });

  $: markdownCssHref = currentTheme === 'dark' ? darkCssUrl : lightCssUrl;
  const source = emojify(markdownSample);

  onDestroy(() => unsubscribeTheme());
</script>

<svelte:head>
  <title>Terralist — Markdown test</title>
  <link rel="stylesheet" href={markdownCssHref} />
</svelte:head>

<Navbar />

<main class="mt-36 mx-4 lg:mt-14 lg:mx-10 text-slate-600 dark:text-slate-200">
  <section class="mt-20 lg:mx-20 flex flex-col gap-4">
    <header class="flex flex-col gap-2">
      <h1 class="text-2xl font-bold tracking-tight dark:text-white">
        Markdown rendering test
      </h1>
      <p class="text-sm">
        Visual fixture for the GFM renderer. Compare against the same source
        rendered on github.com to spot regressions.
      </p>
    </header>
    <article class="flex flex-col gap-4">
      <Markdown {source} />
    </article>
  </section>
</main>
