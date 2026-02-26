<script lang="ts">
  import { onDestroy } from 'svelte';
  import { link } from 'svelte-spa-router';
  import { clickOutside } from 'svelte-use-click-outside';

  import Device from 'svelte-system-info';

  import Input from './/Input.svelte';
  import KeyboardAction from './KeyboardAction.svelte';
  import Icon from './Icon.svelte';

  import { Artifacts, type Artifact } from '@/api/artifacts';
  import { useQuery } from '@/lib/hooks';
  import { computeArtifactUrl } from '@/lib/artifact';

  let open: boolean = false;

  let query: string = '';

  let searchbar: Input | null = null;
  let searchEntries: (HTMLAnchorElement | null)[] = Array.from(
    { length: 10 },
    () => null
  );
  let selectedSearchEntry: number = 0;

  const result = useQuery(Artifacts.getAll);

  let artifacts: Artifact[] = [];
  let filteredArtifacts: Artifact[] = [];

  const unsubscribe = result.subscribe(({ data, isLoading, error }) => {
    if (isLoading || error) {
      return;
    }

    artifacts = data ?? [];
  });

  const useMetaKey = ['macOS', 'iPadOS', 'iOS'].includes(Device.OSName);

  const filterArtifacts = () => {
    const sanitizer: (s: string | undefined) => string = s =>
      s ? s.toLowerCase().replace(/\s+/g, '') : '';

    filteredArtifacts = artifacts
      .filter(({ fullName }) => sanitizer(fullName).includes(sanitizer(query)))
      .filter((_, i) => i < 10);

    selectedSearchEntry = -1;
  };

  const triggerSearchbar = () => {
    filterArtifacts();
    searchbar?.focus();
    open = true;
  };

  const moveSelector = (operator: -1 | 1) => {
    selectedSearchEntry = Math.min(
      Math.max(selectedSearchEntry + operator, 0),
      filteredArtifacts.length - 1
    );

    searchEntries[selectedSearchEntry]?.focus();
  };

  const moveSelectorUp = () => {
    moveSelector(-1);
  };

  const moveSelectorDown = () => {
    moveSelector(1);
  };

  const openEntry = () => {
    searchEntries[selectedSearchEntry]?.click();
  };

  const escapeSearchbar = () => {
    open = false;
    searchbar?.blur();
  };

  onDestroy(() => {
    unsubscribe();
  });
</script>

{#if open}
  <KeyboardAction
    trigger="ArrowUp"
    action={moveSelectorUp}
    preventDefault={true} />
  <KeyboardAction
    trigger="ArrowDown"
    action={moveSelectorDown}
    preventDefault={true} />
  <KeyboardAction trigger="Enter" action={openEntry} preventDefault={true} />
  <KeyboardAction
    trigger="Escape"
    action={escapeSearchbar}
    preventDefault={true} />
{:else if useMetaKey}
  <KeyboardAction
    trigger="Meta+/"
    action={triggerSearchbar}
    preventDefault={true} />
{:else}
  <KeyboardAction
    trigger="Control+/"
    action={triggerSearchbar}
    preventDefault={true} />
{/if}

<nav class="mt-6 lg:mt-0 lg:justify-self-center lg:w-1/2 relative">
  <div class="w-screen lg:w-full relative">
    <div class="mx-8 lg:mx-0">
      <Input
        placeholder="Search modules or providers ({useMetaKey
          ? `Cmd`
          : `Ctrl`}+/)"
        onClick={triggerSearchbar}
        onInput={triggerSearchbar}
        bind:value={query}
        bind:this={searchbar}>
        <Icon
          name="search"
          width="2rem"
          class="absolute mt-3 md:fill-slate-400" />
      </Input>
    </div>
    {#if open}
      <div
        use:clickOutside={escapeSearchbar}
        class="w-10/12 lg:w-full inset-x-0 mx-auto absolute top-12 flex flex-col justify-start bg-white list-none py-2 rounded-lg shadow-md bg-zinc-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200">
        {#if $result.isLoading}
          <div class="py-1 px-5">Loading...</div>
        {:else if $result.error}
          <div class="py-1 px-5">{$result.error}</div>
        {:else if filteredArtifacts.length === 0 && query !== ''}
          <div class="py-1 px-5">0 results found.</div>
        {/if}

        {#each filteredArtifacts as artifact, index (artifact.id)}
          <a
            on:click={escapeSearchbar}
            href={computeArtifactUrl(artifact)}
            bind:this={searchEntries[index]}
            use:link
            tabindex={index}
            class="
              relative
              cursor-pointer
              hover:bg-teal-700
              hover:text-white
              hover:fill-white
              focus:bg-teal-700
              focus:text-white
              focus:fill-white
              select-none
              py-1
              px-2
              flex
              flex-row
              justify-between
            ">
            <span class="block truncate hover:underline focus:underline">
              {artifact.fullName}
            </span>
            {#if artifact.type === 'provider'}
              <div
                class="absolute top-1 right-2 flex justify-between gap-1 items-center fill-inherit">
                <span class="text-xs">(provider)</span>
                <Icon name="cloud" class="fill-inherit" />
              </div>
            {/if}
            {#if artifact.type === 'module'}
              <div
                class="absolute top-1 right-2 flex justify-between gap-1 items-center fill-inherit">
                <span class="text-xs">(module)</span>
                <Icon name="tools" class="fill-inherit" />
              </div>
            {/if}
          </a>
        {/each}
      </div>
    {/if}
  </div>
</nav>
