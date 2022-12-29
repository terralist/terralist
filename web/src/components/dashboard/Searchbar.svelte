<script lang="ts">
  import { onMount } from 'svelte';
  import { clickOutside } from 'svelte-use-click-outside';

  import { fetchArtifacts, type Artifact } from '../../api/artifacts';

  let open: boolean = false;

  let query: String = '';

  let searchbar: HTMLInputElement;
  let searchEntries: HTMLLIElement[] = Array.from({ length: 10 }, () => null);
  let selectedSearchEntry: number = 0;

  let artifacts: Artifact[] = [];
  let filteredArtifacts: Artifact[] = [];

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

  const redirectToArtifact = (id: Number) => {
    open = false;

    // TODO: Perform the actual redirect
    console.log('Open page for artifact with ID', id);
  };

  const onKeyDown = (e: KeyboardEvent) => {
    switch (e.key) {
      case '/':
        e.preventDefault();

        triggerSearchbar();

        break;
    }

    if (open) {
      switch (e.key) {
        case 'ArrowDown':
        case 'ArrowUp':
          e.preventDefault();

          let operator: number = e.key == 'ArrowUp' ? -1 : 1;

          selectedSearchEntry = Math.min(Math.max(selectedSearchEntry + operator, 0), filteredArtifacts.length - 1);
          searchEntries[selectedSearchEntry].focus();

          break;

        case 'Enter':
          e.preventDefault();

          searchEntries[selectedSearchEntry].click();

          break;

        case 'Escape':
          e.preventDefault();

          open = false;
          searchbar.blur();

          break;
      }
    }
  };

  onMount(() => {
    artifacts = fetchArtifacts();
  });
</script>

<svelte:window on:keydown={onKeyDown} />

<nav 
  class="mt-6 lg:mt-0 lg:justify-self-center lg:w-1/2 relative" 
>
  <div class="w-screen lg:w-full relative">
    <div class="mx-8 lg:mx-0">
      <i class="absolute fa fa-search px-2 py-3 text-center text-slate-400"></i>
      <input type="text"
        class="w-full h-10 pl-8 pr-0.5 bg-zinc-100 dark:bg-zinc-700 text-slate-800 dark:text-slate-200 shadow border-none text-sm rounded-lg focus:ring-0 outline-none"
        placeholder="Search modules or providers (/)"
        bind:this={searchbar}
        bind:value={query}
        on:click={triggerSearchbar}
        on:input={triggerSearchbar}
      />
    </div>
    {#if open}
      <div
        use:clickOutside={() => {open = false}}
        class="w-10/12 lg:w-full inset-x-0 mx-auto absolute top-12 flex flex-col justify-start bg-white list-none py-2 rounded-lg shadow-md bg-zinc-100 dark:bg-zinc-700 text-slate-800 dark:text-slate-200"
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