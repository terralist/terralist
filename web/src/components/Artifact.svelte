<script lang="ts">
  import { onDestroy } from 'svelte';
  import { push } from 'svelte-spa-router';
  import SvelteMarkdown from 'svelte-markdown';

  import lightHref from 'github-markdown-css/github-markdown-light.css?url';
  import darkHref from 'github-markdown-css/github-markdown-dark.css?url';
  import hlLightHref from 'highlight.js/styles/github.css?url';
  import hlDarkHref from 'highlight.js/styles/github-dark.css?url';

  import config from '@/config';
  import { indent } from '@/lib/utils';
  import { useQuery } from '@/lib/hooks';
  import context, { type Theme } from '@/context';

  import Icon from './Icon.svelte';
  import Dropdown from './Dropdown.svelte';
  import FullPageError from './FullPageError.svelte';
  import LoadingScreen from './LoadingScreen.svelte';
  import MarkdownCode from './MarkdownCode.svelte';

  import {
    Artifacts,
    type ArtifactVersion,
    type ArtifactVersionWithDocumentation
  } from '@/api/artifacts';
  import { computeArtifactUrl, type LocatableArtifact } from '@/lib/artifact';

  let currentTheme: Theme = 'light';

  const themeUnsubscribe = context.theme.subscribe(theme => {
    currentTheme = theme;
  });

  export let type: 'module' | 'provider';
  export let namespace: string;
  export let name: string;
  export let provider: string = '';
  export let version: string = '';

  const moduleTemplate = `
    module "${name}" {
      source  = "${config.runtime.TERRALIST_CANONICAL_DOMAIN}/${namespace}/${name}/${provider}"
      version = "${version}"
    }
  `;

  const providerTemplate = `
    terraform {
      required_providers {
        ${name} = {
          source = "${config.runtime.TERRALIST_CANONICAL_DOMAIN}/${namespace}/${name}"
          version = "${version}"
        }
      }
    }

    provider "${name}" {
      # Configuration options
    }
  `;

  const template = indent({
    s: type === 'module' ? moduleTemplate : providerTemplate,
    n: 4,
    reverse: true
  });

  const onOptionSelect = (option: string) => {
    const url = computeArtifactUrl({
      type: type,
      namespace: namespace,
      name: name,
      provider: type == 'module' ? provider : undefined,
      version: option
    } as LocatableArtifact);
    push(url);
  };

  let label: string = version;

  const result = useQuery<ArtifactVersion[]>(
    Artifacts.getAllVersionsForOne,
    namespace,
    name,
    provider
  );

  let versions: string[] = [];

  const unsubscribe = result.subscribe(res => {
    if (res.error || res.isLoading) {
      return;
    }

    versions = res.data ?? [];

    if (versions.length == 0) {
      return;
    }

    // If there is no version, or user selected 'latest' version
    if (!version || version == 'latest') {
      version = versions[0];
    }

    // If the selected version is the latest, change the label
    if (version == versions[0]) {
      label = `${version} (latest)`;
    }
  });

  let documentation: string | undefined;

  const versionUnsubscribe = useQuery<ArtifactVersionWithDocumentation>(
    Artifacts.getOneVersion,
    namespace,
    name,
    provider,
    version
  ).subscribe(res => {
    if (res.isLoading) {
      return;
    }

    if (res.error) {
      console.error(
        'Skipping module version display as it cannot be fetched:',
        res.error
      );
      return;
    }

    if (res.data) {
      documentation = res.data.documentation;
    }
  });

  onDestroy(() => {
    unsubscribe();
    versionUnsubscribe();
    themeUnsubscribe();
  });
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href={currentTheme == 'light' ? lightHref : darkHref} />
  <link
    rel="stylesheet"
    href={currentTheme == 'light' ? hlLightHref : hlDarkHref} />
</svelte:head>
<main class="mt-36 mx-4 lg:mt-14 lg:mx-10 text-slate-600 dark:text-slate-200">
  {#if $result.isLoading}
    <LoadingScreen />
  {:else if $result.error}
    <FullPageError code={0} message={$result.error} />
  {:else if !versions.includes(version)}
    <FullPageError
      code={404}
      message="This artifact version does not currently exist on the server." />
  {:else}
    <section class="mt-20 lg:mx-20 flex flex-col gap-8">
      <div
        class="flex flex-col lg:flex-row justify-between items-start gap-8 lg:items-center">
        <div class="flex justify-start items-center gap-10 mb-4">
          <div
            class="flex flex-col justify-center items-center dark:text-white">
            <Icon
              name={type === 'provider' ? 'cloud' : 'tools'}
              width="8rem"
              height="8rem" />
            <span class="text-xs text-zinc-800 dark:text-white">
              {`(${type})`}
            </span>
          </div>
          <div class="flex flex-col justify-center items-start">
            <h2
              class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white break-words">
              {name + (type === 'module' ? ` (${provider})` : '')}
            </h2>
            <h3 class="text-zinc-800 dark:text-zinc-100">
              @{namespace}
            </h3>
          </div>
        </div>
        <div class="w-full lg:w-auto">
          <Dropdown {label} options={versions} onSelect={onOptionSelect} />
        </div>
      </div>
      <div
        class="bg-gray-100 dark:bg-gray-800 border border-teal-400 dark:border-teal-600 p-4 flex flex-col gap-4">
        <h2 class="text-lg font-bold">Usage</h2>
        <p class="text-xs">
          Copy and paste into your Terraform configuration, insert the
          variables, and run <code>terraform init</code>:
        </p>
        <pre
          class="bg-gray-200 dark:bg-gray-700 border border-slate-400 dark:border-slate-600 p-3 text-xs overflow-y-auto">{template}</pre>
      </div>
      {#if documentation}
        <div class="m-6 p-4 flex flex-col gap-4">
          <h2 class="text-lg font-bold">Readme</h2>
          <div class="markdown-body bg-slate-50">
            <SvelteMarkdown
              source={documentation}
              renderers={{ code: MarkdownCode }} />
          </div>
        </div>
      {/if}
    </section>
  {/if}
</main>
