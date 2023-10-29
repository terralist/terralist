<script lang="ts">
  import { link } from "svelte-spa-router";
  
  import Icon from "./Icon.svelte";

  import type { Artifact } from "@/api/artifacts";
  import { timeSince } from "@/lib/utils";
  import { computeArtifactUrl } from '@/lib/artifact';

  export let artifact: Artifact;

  const url = computeArtifactUrl(artifact);
</script>

<section class="p-6 w-full bg-white rounded-lg border border-gray-200 shadow-md dark:bg-gray-800 dark:border-gray-700">
  <div class="flex justify-between items-center mb-4">
    <div class="flex flex-col justify-center items-start">
      <a href={url} use:link>
        <h2 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white break-words">
          {artifact.name + (artifact.type === 'module' ? ` (${artifact.provider})` : '')}
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
        height="1.5rem"
      />
      <span class="text-xs text-zinc-800 dark:text-white">
        {`(${artifact.type})`}
      </span>
    </div>
  </div>
  <div class="grid grid-cols-2 gap-4 mb-3 font-normal text-gray-700 dark:text-gray-400 text-sm">
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
    use:link
  >
    View documentation
    <Icon name="arrow-forward" class="ml-2 -mr-1" />
  </a>
</section>
