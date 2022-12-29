<script lang="ts">
  import type { Artifact } from "../../../api/artifacts";

  export let artifact: Artifact;
</script>

{#key artifact.id}
  <section class="p-6 w-full bg-white rounded-lg border border-gray-200 shadow-md dark:bg-gray-800 dark:border-gray-700">
    <div class="flex justify-between items-center mb-4">
      <div class="flex flex-col justify-center items-start">
        <!-- svelte-ignore a11y-invalid-attribute -->
        <a href="#">
          <h2 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white break-words">
            {artifact.name + (artifact.type === 'module' ? ` (${artifact.provider})` : '')}
          </h2>
        </a>
        <h3 class="text-zinc-800 dark:text-zinc-100">
          @{artifact.authority}
        </h3>
      </div>
      <div class="flex flex-col justify-center items-center dark:text-white">
        <i class="fa {artifact.type === 'provider' ? 'fa-cloud' : 'fa-hammer'}"></i>
        <span class="text-xs text-zinc-800 dark:text-white">
          {`(${artifact.type})`}
        </span>
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4 mb-3 font-normal text-gray-700 dark:text-gray-400 text-sm">
      <p class="place-self-start">Version:</p>
      <p class="place-self-end">1.2.1</p>
      <p class="place-self-start">Updated:</p>
      <p class="place-self-end">2 days ago</p>
      <p class="place-self-start">Published:</p>
      <p class="place-self-end">1 year ago</p>
    </div>
    <!-- svelte-ignore a11y-invalid-attribute -->
    <a href="#"
      class="w-full inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center text-slate-50 bg-teal-400 rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      View documentation
      <svg aria-hidden="true" class="ml-2 -mr-1 w-4 h-4" fill="currentColor" viewBox="0 0 20 20"
        xmlns="http://www.w3.org/2000/svg">
        <path fill-rule="evenodd"
          d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z"
          clip-rule="evenodd"></path>
      </svg>
    </a>
  </section>
{/key}