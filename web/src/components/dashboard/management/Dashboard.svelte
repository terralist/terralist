<script lang="ts">
  import { type Authority as AuthorityT, fetchAuthorities } from "../../../api/authorities";

  import Authority from "./Authority.svelte";

  let authorities: AuthorityT[] = fetchAuthorities();
</script>

<main class="mt-36 mx-4 lg:mt-14 lg:mx-10">
  <section class="flex">
    <button class="mt-4 inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center text-slate-50 bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      <i class="fa fa-plus pt-0.5 w-4 h-4"></i>
      <span class="ml-2 text-sm text-light uppercase">
        New Authority
      </span>
    </button>
  </section>
  <section class="mt-8 lg:mx-20">
    <div class="w-full p-2 px-6 grid grid-cols-6 lg:grid-cols-10 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span class="col-span-2 lg:col-span-6">
        Name
      </span>
      <span>
        Policy
      </span>
      <span>
        Signing Keys
      </span>
      <span>
        API Keys
      </span>
      <span class="place-self-end">
        Actions
      </span>
    </div>
    {#each authorities as authority}
      {#key authority.id}
        <Authority authority={authority} />
      {/key}
    {/each}
  </section>
</main>