<script lang="ts">
  import { link } from 'svelte-spa-router';

  import Icon from './Icon.svelte';

  import type { Artifact } from '@/api/artifacts';
  import { timeSince } from '@/lib/utils';
  import { computeArtifactUrl } from '@/lib/artifact';

  export let artifact: Artifact;
  export let variant: 'grid' | 'list' = 'grid';

  const url = computeArtifactUrl(artifact);
</script>

{#if variant === 'grid'}
  <!-- Grid Card (original) -->
  <section
    class="p-6 w-full bg-white rounded-lg border border-gray-200 shadow-md dark:bg-gray-800 dark:border-gray-700">
    <div class="flex justify-between items-center mb-4">
      <div class="flex flex-col justify-center items-start">
        <a href={url} use:link>
          <h2
            class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white break-words">
            {artifact.name +
              (artifact.type === 'module' ? ` (${artifact.provider})` : '')}
          </h2>
        </a>
        <h3 class="text-zinc-800 dark:text-zinc-100">
          @{artifact.namespace}
        </h3>
      </div>
      <div class="flex flex-col justify-center items-center dark:text-white">
        <Icon
          name={artifact.type === 'provider' ? 'cloud' : 'tools'}
          width="1.5rem"
          height="1.5rem" />
        <span class="text-xs text-zinc-800 dark:text-white">
          {`(${artifact.type})`}
        </span>
      </div>
    </div>
    <div
      class="grid grid-cols-2 gap-4 mb-3 font-normal text-gray-700 dark:text-gray-400 text-sm">
      <p class="place-self-start">Version:</p>
      <p class="place-self-end">{artifact.versions[0]}</p>
      <p class="place-self-start">Updated:</p>
      <p class="place-self-end">{timeSince(artifact.updatedAt)}</p>
      <p class="place-self-start">Published:</p>
      <p class="place-self-end">{timeSince(artifact.createdAt)}</p>
    </div>
    <a
      class="
        w-full
        inline-flex
        justify-center
        items-center
        py-2
        px-3
        text-sm
        font-medium
        text-center
        text-slate-600
        dark:text-slate-200
        bg-teal-400
        rounded-lg
        hover:bg-teal-500
        focus:ring-4
        focus:outline-none
        focus:ring-green-300
        dark:bg-teal-700
        dark:hover:bg-teal-800
        dark:focus:ring-green-700
        {$$props.class}
      "
      href={url}
      use:link>
      View documentation
      <Icon name="arrow-forward" class="ml-2 -mr-1" />
    </a>
  </section>
{:else}
  <!-- List Row (compact) -->
  <a
    href={url}
    use:link
    class="flex items-center gap-4 p-4 bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md hover:border-teal-300 dark:bg-gray-800 dark:border-gray-700 dark:hover:border-teal-600 transition-all {$$props.class}">
    <div
      class="flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-700">
      <Icon
        name={artifact.type === 'provider' ? 'cloud' : 'tools'}
        width="1.25rem"
        height="1.25rem" />
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2">
        <h3
          class="text-lg font-semibold text-gray-900 dark:text-white truncate">
          {artifact.name}{artifact.type === 'module'
            ? ` (${artifact.provider})`
            : ''}
        </h3>
        <span
          class="px-2 py-0.5 text-xs font-medium rounded-full bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300">
          {artifact.type}
        </span>
      </div>
      <p class="text-sm text-gray-500 dark:text-gray-400 truncate">
        @{artifact.namespace}
      </p>
    </div>
    <div
      class="hidden sm:flex items-center gap-6 text-sm text-gray-500 dark:text-gray-400">
      <div class="text-center">
        <p class="font-medium text-gray-900 dark:text-white">
          {artifact.versions[0]}
        </p>
        <p class="text-xs">version</p>
      </div>
      <div class="text-center">
        <p class="font-medium text-gray-900 dark:text-white">
          {timeSince(artifact.updatedAt)}
        </p>
        <p class="text-xs">updated</p>
      </div>
    </div>
    <Icon
      name="arrow-forward"
      width="1.25rem"
      height="1.25rem"
      class="text-gray-400" />
  </a>
{/if}
