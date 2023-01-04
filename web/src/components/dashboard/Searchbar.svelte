<script lang="ts">
  import { onMount } from 'svelte';
  import { clickOutside } from 'svelte-use-click-outside';
  import Device from 'svelte-system-info';

  import { fetchArtifacts, type Artifact } from '../../api/artifacts';
  import Input from '../Input.svelte';
  import KeyboardAction from '../KeyboardAction.svelte';

  let open: boolean = false;

  let query: string = "";

  let searchbar: any;
  let searchEntries: HTMLLIElement[] = Array.from({ length: 10 }, () => null);
  let selectedSearchEntry: number = 0;

  let artifacts: Artifact[] = [];
  let filteredArtifacts: Artifact[] = [];

  const useMetaKey = ["macOS", "iPadOS", "iOS"].includes(Device.OSName);

  const filterArtifacts = () => {
    const sanitizer: (s: String | undefined) => string = (
      (s) => s ? s.toLowerCase().replace(/\s+/g, '') : ""
    );
    
    filteredArtifacts = artifacts.
      filter(({ fullName }) => sanitizer(fullName).includes(sanitizer(query))).
      filter((_, i) => i < 10);
    
    selectedSearchEntry = -1;
  }

  const triggerSearchbar = () => {
    filterArtifacts();
    searchbar.focus();
    open = true;
  };

  const redirectToArtifact = (id: string) => {
    open = false;

    // TODO: Perform the actual redirect
    console.log('Open page for artifact with ID', id);
  };

  const moveSelector = (operator: -1 | 1) => {
    selectedSearchEntry = Math.min(Math.max(selectedSearchEntry + operator, 0), filteredArtifacts.length - 1);

    searchEntries[selectedSearchEntry].focus();
  };


  const moveSelectorUp = () => { moveSelector(-1); };

  const moveSelectorDown = () => { moveSelector(1); }

  const openEntry = () => {
    searchEntries[selectedSearchEntry].click();
  };

  const escapeSearchbar = () => {
    open = false;
    searchbar.blur();
  };

  onMount(() => {
    artifacts = fetchArtifacts();
  });
</script>

{#if open}
  <KeyboardAction trigger={'ArrowUp'} action={moveSelectorUp} preventDefault={true} />
  <KeyboardAction trigger={'ArrowDown'} action={moveSelectorDown} preventDefault={true} />
  <KeyboardAction trigger={'Enter'} action={openEntry} preventDefault={true} />
  <KeyboardAction trigger={'Escape'} action={escapeSearchbar} preventDefault={true} />
{:else}
  {#if useMetaKey}
    <KeyboardAction trigger={'Meta+/'} action={triggerSearchbar} preventDefault={true} />
  {:else}
    <KeyboardAction trigger={'Control+/'} action={triggerSearchbar} preventDefault={true} />
  {/if}
{/if}

<nav 
  class="mt-6 lg:mt-0 lg:justify-self-center lg:w-1/2 relative" 
>
  <div class="w-screen lg:w-full relative">
    <div class="mx-8 lg:mx-0">
      <Input 
        placeholder="Search modules or providers ({useMetaKey ? `Cmd` : `Ctrl`}+/)"
        onClick={triggerSearchbar}
        onInput={triggerSearchbar}
        bind:value={query}
        bind:this={searchbar}
      >
        <i class="absolute fa fa-search px-2 py-3 text-center text-slate-400"></i>      
      </Input>
    </div>
    {#if open}
      <div
        use:clickOutside={() => {open = false}}
        class="w-10/12 lg:w-full inset-x-0 mx-auto absolute top-12 flex flex-col justify-start bg-white list-none py-2 rounded-lg shadow-md bg-zinc-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200"
      >
        {#if filteredArtifacts.length === 0 && query !== ''}
          <div class="py-1 px-5">0 results found.</div>
        {/if}

        {#each filteredArtifacts as artifact, index}
          {#key artifact.id}
            <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <li 
              on:click={() => redirectToArtifact(artifact.id)}
              bind:this={searchEntries[index]}
              tabindex="{index}"
              class="relative cursor-pointer hover:bg-teal-700 hover:text-white focus:bg-teal-700 focus:text-white select-none py-1 px-2 flex flex-row justify-between">
              <span class="block truncate hover:underline focus:underline">
                {artifact.fullName}
              </span>
              {#if artifact.type === 'provider'}
                <div class="absolute top-1 right-2">
                  <span class="text-xs">(provider)</span>
                  <i class="fa fa-cloud"></i>
                </div>
              {/if}
              {#if artifact.type === 'module'}
                <div class="absolute top-1 right-2">
                  <span class="text-xs">(module)</span>
                  <i class="fa fa-hammer"></i>
                </div>
              {/if}
            </li>
          {/key}
        {/each}
      </div>
    {/if}
  </div>
</nav>