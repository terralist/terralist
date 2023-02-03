<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { writable, type Writable } from "svelte/store";

  import { fetchArtifacts, type Artifact } from "@/api/artifacts";

  import { defaultIfNull } from '@/lib/utils';

  import Icon from "./Icon.svelte";
  import ArtifactCard from "./ArtifactCard.svelte";

  let pagesToDisplay: number = 5;
  let pageCount: number = 10;
  let itemsPerPage: number = 8;
  let pages: number[] = [];
  let currentPage: number = 0;

  let artifacts: Artifact[] = [];

  const filters: Writable<{
    modulesEnabled: boolean,
    providersEnabled: boolean,
  }> = writable({
    modulesEnabled: defaultIfNull(JSON.parse(sessionStorage.getItem('filters.modules')), true),
    providersEnabled: defaultIfNull(JSON.parse(sessionStorage.getItem('filters.providers')), true),
  });

  const updateFilters = () => {
    sessionStorage.setItem('filters.modules', JSON.stringify($filters.modulesEnabled));
    sessionStorage.setItem('filters.providers', JSON.stringify($filters.providersEnabled));
  };

  const initPages = () => {
    const artifactsCount = artifacts.length;

    pageCount = artifactsCount > 0 ? Math.floor(artifactsCount / itemsPerPage + 1) : 0;
    pagesToDisplay = Math.min(pageCount, 5);
  };

  const buildPages = (pageIndex: number) => {
    currentPage = pageIndex;
    updateArtifacts();

    let start: number = 0,
      end: number = pagesToDisplay,
      leftMid: number = (pagesToDisplay - 1) / 2,
      rightMid: number = (pagesToDisplay + 1) / 2;

    if (pageIndex > leftMid) {
      start = pageIndex - leftMid;
      end = start + pagesToDisplay;
    }

    if (pageIndex > pageCount - rightMid) {
      start = pageCount - pagesToDisplay;
      end = pageCount;
    }

    pages = Array.from({ length: end - start }, (_, i) => start + i);
  };

  const updateArtifacts = () => {
    const currentFilters = (
      $filters.modulesEnabled ? ['module'] : []
    ).concat(
      $filters.providersEnabled ? ['provider'] : []
    );

    artifacts = fetchArtifacts().
      filter((artifact: Artifact) => currentFilters.includes(artifact.type)).
      filter((_, id) => (id >= itemsPerPage * currentPage && id < itemsPerPage * (1 + currentPage)));
  };

  let filtersUnsubscribe: () => void;

  onMount(() => {
    updateArtifacts();

    initPages();

    filtersUnsubscribe = filters.subscribe(() => {
      updateFilters();
      updateArtifacts();

      initPages();
      buildPages(0);
    });
  });

  onDestroy(() => {
    filtersUnsubscribe();
  });
</script>

<main class="mt-36 lg:mt-20 mx-10">
  <div class="flex justify-center items-center">
    <div class="flex flex-row">
      <input 
        id="modules-checkbox"
        type="checkbox"
        class="mt-0.5 w-4 h-4 text-blue-600 bg-gray-100 rounded border-gray-300 dark:bg-gray-700 dark:border-gray-600"
        value="module"
        bind:checked={$filters.modulesEnabled}
      />
      <label for="modules-checkbox" class="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300">
        Modules
      </label>
    </div>
    <div class="ml-4 flex flex-row">
      <input 
        id="providers-checkbox" 
        type="checkbox"
        class="mt-0.5 w-4 h-4 text-blue-600 bg-gray-100 rounded border-gray-300 dark:bg-gray-700 dark:border-gray-600"
        value="provider"
        bind:checked={$filters.providersEnabled} 
      />
      <label for="providers-checkbox" class="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300">
        Providers
      </label>
    </div>
  </div>

  {#if pageCount > 0}
    <div class="mt-4 flex flex-col justify-center items-center sm:grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {#each artifacts as artifact}
        <ArtifactCard artifact={artifact}/>
      {/each}
    </div>
  {/if}

  {#if pageCount > 0}
    <div class="flex gap-1 my-8 justify-center items-center">
      <button
        class="grid place-items-center w-8 h-8 p-0 border-0 rounded cursor-pointer bg-slate-200 text-zinc-800 dark:bg-slate-800 dark:text-slate-200 {currentPage === 0 ? 'opacity-25 -z-10' : 'opacity-100'}"
        on:click={() => buildPages(0)}
        disabled={currentPage === 0 ? true : false}
      >
        <Icon name="arrow-left" width="1.25rem" height="1.25rem" />
      </button>

      {#each pages as page}
        <button
          class="grid place-items-center w-8 h-8 p-0 border-0 rounded cursor-pointer bg-slate-200 text-zinc-800 dark:bg-slate-800 dark:text-slate-200 {currentPage === page ? 'bg-teal-300 dark:bg-teal-800' : ''}"
          on:click={() => buildPages(page)} 
        >
          {page + 1}
        </button>
      {/each}

      <button
        class="grid place-items-center w-8 h-8 p-0 border-0 rounded cursor-pointer bg-slate-200 text-zinc-800 dark:bg-slate-800 dark:text-slate-200 {currentPage === (pageCount - 1) ? 'opacity-25' : 'opacity-100'}"
        on:click={() => buildPages(pageCount - 1)}
        disabled={currentPage === (pageCount - 1) ? true : false}
      >
        <Icon name="arrow-right" width="1.25rem" height="1.25rem" />
      </button>
    </div>
  {/if}

  {#if pageCount === 0}
    <div 
      class="absolute top-0 left-0 flex justify-center items-center text-center w-screen h-screen -z-10"
    >
      <p class="text-zinc-900 dark:text-zinc-100">There's nothing to see here.</p>
    </div>
  {/if}
</main>