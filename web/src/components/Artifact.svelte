<script lang="ts">
  import { onDestroy } from 'svelte';
  import { push } from 'svelte-spa-router';

  import context, { type Theme } from '@/context';

  let currentTheme: Theme = 'light';
  const unsubscribeTheme = context.theme.subscribe(t => {
    currentTheme = t;
  });

  import config from '@/config';
  import { indent } from '@/lib/utils';
  import { useFlag, useQuery } from '@/lib/hooks';

  import Icon from './Icon.svelte';
  import Button from './Button.svelte';
  import ConfirmationModal from './ConfirmationModal.svelte';
  import ErrorModal from './ErrorModal.svelte';
  import FormModal from './FormModal.svelte';
  import Dropdown from './Dropdown.svelte';
  import FullPageError from './FullPageError.svelte';
  import LoadingScreen from './LoadingScreen.svelte';
  import Markdown from './markdown/Markdown.svelte';
  import { emojify } from 'node-emoji';

  import {
    Artifacts,
    type ArtifactVersions,
    type ArtifactVersionWithDocumentation,
    type Submodule,
    type VersionDetails
  } from '@/api/artifacts';
  import { Authorities } from '@/api/authorities';
  import { Registry, type OfferedVersion } from '@/api/registry';
  import { Rules } from '@/api/rules';
  import type { FormEntry } from '@/lib/form';
  import cmp from 'semver-compare';
  import { computeArtifactUrl, type LocatableArtifact } from '@/lib/artifact';

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

  const result = useQuery<ArtifactVersions>(
    Artifacts.getAllVersionsForOne,
    namespace,
    name,
    provider
  );

  let versions: string[] = [];
  let details: VersionDetails[] = [];

  let canDelete = false;
  let canFetch = false;
  let canBlock = false;
  // offered lists the upstream versions Terralist does not hold yet.
  let offered: OfferedVersion[] = [];
  let actionError = '';

  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);
  const [blockModalEnabled, showBlockModal, hideBlockModal] = useFlag(false);
  const [fetchModalEnabled, showFetchModal, hideFetchModal] = useFlag(false);

  // badgesOf names what sets a version apart from one uploaded with its
  // signature material.
  const badgesOf = (v: VersionDetails | undefined): string[] =>
    [
      v?.origin === 'upstream' ? 'pulled from upstream' : '',
      v?.mirrorOnly ? 'network mirror only' : ''
    ].filter(b => b);

  $: badges = badgesOf(details.find(d => d.version === version));
  // The version selector is narrow, so it tags versions in short.
  $: versionLabels = Object.fromEntries(
    details.map(d => [
      d.version,
      [
        d.version,
        d.origin === 'upstream' ? 'upstream' : '',
        d.mirrorOnly ? 'mirror only' : ''
      ]
        .filter(p => p)
        .join(' · ')
    ])
  );

  const unsubscribe = result.subscribe(res => {
    if (res.error || res.isLoading) {
      return;
    }

    details = res.data?.versions ?? [];
    versions = details.map(d => d.version);
    canDelete = res.data?.canDelete ?? false;
    canFetch = res.data?.canFetch ?? false;
    canBlock = res.data?.canBlock ?? false;

    if (canFetch) {
      loadOffered();
    }

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

  const loadOffered = async () => {
    const res = await Registry.offeredVersions(namespace, name, provider);
    if (res.status !== 'OK') {
      return;
    }

    offered = res.data
      .filter(o => !versions.includes(o.version))
      .sort((a, b) => cmp(b.version, a.version));
  };

  // leave moves away from a version that no longer exists: to the latest
  // remaining one, or to the dashboard when none is left.
  const leave = () => {
    const remaining = versions.filter(v => v !== version);
    if (remaining.length > 0) {
      onOptionSelect(remaining[0]);
    } else {
      push('/');
    }
  };

  const deleteVersion = async () => {
    const res = await Artifacts.delete(namespace, name, provider, version);
    if (res.status !== 'OK') {
      actionError = res.message;
      return;
    }

    leave();
  };

  // blockVersion denies the version by a rule of the authority, so it is not
  // pulled again, then deletes the pulled copy.
  const blockVersion = async () => {
    const authorities = await Authorities.getAll();
    const authority =
      authorities.status === 'OK'
        ? authorities.data.find(
            a => a.name.toLowerCase() === namespace.toLowerCase()
          )
        : undefined;
    if (!authority) {
      actionError = `Could not find the authority ${namespace}.`;
      return;
    }

    const rule = await Rules.create(authority.id, {
      kind: type,
      name: type === 'module' ? `${name}/${provider}` : name,
      version,
      effect: 'deny'
    });
    if (rule.status !== 'OK') {
      actionError = rule.message;
      return;
    }

    await deleteVersion();
  };

  const fetchVersion = async (
    entries: Map<string, string | string[] | undefined>
  ) => {
    const value = (id: string) => {
      const entry = entries.get(id);
      return (Array.isArray(entry) ? entry.at(0) : entry) ?? '';
    };

    const target = value('fetchVersion');
    const requested = value('fetchPlatforms')
      .split(',')
      .map(p => p.trim())
      .filter(p => p);
    const platforms =
      requested.length > 0
        ? requested
        : (offered.find(o => o.version === target)?.platforms ?? []).map(
            p => `${p.os}_${p.arch}`
          );

    const res = await Artifacts.fetchFromUpstream(
      namespace,
      name,
      provider,
      target,
      platforms
    );
    if (res.status !== 'OK') {
      actionError = res.message;
      return;
    }

    const failed = (res.data.results ?? []).filter(r => r.error);
    if (failed.length > 0) {
      actionError = failed.map(r => `${r.platform}: ${r.error}`).join('\n');
      return;
    }

    onOptionSelect(target);
  };

  $: selected = details.find(d => d.version === version);

  let fetchEntries: FormEntry[] = [];
  $: fetchEntries = [
    {
      id: 'fetchVersion',
      name: 'Version',
      type: 'select',
      value: offered[0]?.version,
      options: offered.map(o => ({ value: o.version, label: o.version }))
    },
    ...(type === 'provider'
      ? [
          {
            id: 'fetchPlatforms',
            name: 'Platforms',
            type: 'text',
            placeholder: 'All platforms, or linux_amd64,darwin_arm64'
          } satisfies FormEntry
        ]
      : [])
  ];

  let documentation: string | undefined;
  let submodules: Submodule[] = [];
  let selectedSubmodule: string | null = null;
  let submoduleLabel: string = 'Select a submodule';
  let submoduleDocumentation: string | undefined;

  // Cache for submodule documentation to avoid redundant API calls
  let submoduleDocsCache: Map<string, string> = new Map();

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
      documentation = emojify(res.data.documentation || '');
      submodules = res.data.submodules || [];
      // Reset submodule selection when version changes
      selectedSubmodule = null;
      submoduleLabel = 'Select a submodule';
      submoduleDocumentation = undefined;
      // Clear cache when version changes to avoid stale data
      submoduleDocsCache.clear();
    }
  });

  const onSubmoduleSelect = async (submodulePath: string) => {
    if (type !== 'module') return;

    selectedSubmodule = submodulePath;
    submoduleLabel = submodulePath;

    // Check cache first to avoid redundant API calls
    if (submoduleDocsCache.has(submodulePath)) {
      submoduleDocumentation = submoduleDocsCache.get(submodulePath);
      return;
    }

    // Set to undefined to show loading state
    submoduleDocumentation = undefined;

    try {
      const result = await Artifacts.getSubmoduleDocumentation(
        namespace,
        name,
        provider,
        version,
        submodulePath
      );

      if (result.status === 'OK' && result.data) {
        const docs = emojify(result.data.documentation || '');
        // Cache the documentation for future use
        submoduleDocsCache.set(submodulePath, docs);
        submoduleDocumentation = docs;
      }
    } catch (error) {
      console.error('Failed to load submodule documentation:', error);
      submoduleDocumentation = undefined;
    }
  };

  onDestroy(() => {
    unsubscribe();
    versionUnsubscribe();
    unsubscribeTheme();
  });

  import lightCssUrl from 'github-markdown-css/github-markdown-light.css?url';
  import darkCssUrl from 'github-markdown-css/github-markdown-dark.css?url';

  $: markdownCssHref = currentTheme === 'dark' ? darkCssUrl : lightCssUrl;
</script>

<svelte:head>
  <link rel="stylesheet" href={markdownCssHref} />
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
            {#if badges.length > 0}
              <div data-testid="version-badges" class="mt-2 flex gap-2">
                {#each badges as badge (badge)}
                  <span
                    class="px-2 rounded-lg text-xs uppercase bg-teal-200 dark:bg-teal-900 text-zinc-800 dark:text-zinc-100">
                    {badge}
                  </span>
                {/each}
              </div>
            {/if}
          </div>
        </div>
        <div class="w-full lg:w-auto">
          <Dropdown
            {label}
            options={versions}
            optionLabels={versionLabels}
            onSelect={onOptionSelect} />
        </div>
      </div>
      {#if (canFetch && offered.length > 0) || canDelete}
        <div class="flex flex-wrap justify-end gap-2">
          {#if canFetch && offered.length > 0}
            <Button class="!w-auto" onClick={showFetchModal}>
              Fetch from upstream
            </Button>
          {/if}
          {#if canBlock && selected?.origin === 'upstream'}
            <Button class="!w-auto" onClick={showBlockModal}>
              Block version
            </Button>
          {/if}
          {#if canDelete}
            <Button class="!w-auto" onClick={showDeleteModal}>
              Delete version
            </Button>
          {/if}
        </div>
      {/if}
      {#if type === 'module' && submodules && submodules.length > 0}
        <div
          class="mt-6 flex flex-col lg:flex-row items-start lg:items-center gap-4">
          <div class="flex items-center gap-4">
            <h2 class="text-lg font-bold">Submodules:</h2>
            <div class="w-80">
              <Dropdown
                label={submoduleLabel}
                options={submodules.map(sm => sm.path)}
                onSelect={onSubmoduleSelect} />
            </div>
          </div>
        </div>
        {#if selectedSubmodule && submoduleDocumentation}
          <div
            class="mt-6 p-4 border border-gray-300 dark:border-gray-600 rounded">
            <h3 class="text-md font-bold mb-2">{selectedSubmodule}</h3>
            <Markdown source={submoduleDocumentation} />
          </div>
        {:else if selectedSubmodule}
          <div
            class="mt-6 p-4 border border-gray-300 dark:border-gray-600 rounded">
            <p class="text-sm text-gray-500 dark:text-gray-400">
              Loading documentation...
            </p>
          </div>
        {/if}
      {/if}
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
        <div class="flex flex-col gap-4">
          <h2 class="text-lg font-bold">Readme</h2>
          <Markdown source={documentation} />
        </div>
      {/if}
    </section>
  {/if}
</main>

<ConfirmationModal
  title={`Delete version ${version}`}
  enabled={$deleteModalEnabled}
  onClose={hideDeleteModal}
  onSubmit={deleteVersion}>
  Version <b>{version}</b> of <b>{namespace}/{name}</b> will be deleted from
  Terralist.
  {#if selected?.origin === 'upstream'}
    It was pulled from the upstream, so it is pulled again on the next request
    for it, unless it is blocked.
  {/if}
  <br /><br />
  Are you sure?
</ConfirmationModal>

<ConfirmationModal
  title={`Block version ${version}`}
  enabled={$blockModalEnabled}
  onClose={hideBlockModal}
  onSubmit={blockVersion}>
  A rule of <b>{namespace}</b> will deny version <b>{version}</b> of
  <b>{name}</b>, so it is no longer pulled from the upstream, and the pulled
  copy will be deleted.
  <br /><br />
  Are you sure?
</ConfirmationModal>

<FormModal
  title={`Fetch a version of ${namespace}/${name} from the upstream`}
  enabled={$fetchModalEnabled}
  onClose={hideFetchModal}
  onSubmit={fetchVersion}
  entries={fetchEntries} />

{#if actionError}
  <ErrorModal bind:message={actionError} />
{/if}
