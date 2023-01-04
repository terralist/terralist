<script lang="ts">
  import { type Authority as AuthorityT, fetchAuthorities } from "../../../api/authorities";
    import Input from "../../Input.svelte";
  import Modal from "../../Modal.svelte";

  import Authority from "./Authority.svelte";

  let authorities: AuthorityT[] = fetchAuthorities();

  let modalOpen: boolean = false;
</script>

<main class="mt-36 mx-4 lg:mt-14 lg:mx-10 text-slate-600 dark:text-slate-200">
  <section class="mt-20 lg:mx-20">
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
    <div class="w-full px-6 grid grid-cols-6 lg:grid-cols-10 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span class="col-span-4 flex">
        <button on:click={() => {modalOpen = true;}} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
          <i class="fa fa-plus pt-0.5 w-4 h-4"></i>
          <span class="ml-2 text-sm text-light uppercase">
            New Authority
          </span>
        </button>
      </span>
    </div>
  </section>

  <Modal title="New authority" enabled={modalOpen} onClose={() => {modalOpen = false;}}>
    <span slot="body">
      <div class="p-2 w-full grid grid-cols-4 gap-4">
        <span class="pt-2">
          Name *
        </span>
        <span class="col-span-3">
          <Input type="text" />
        </span>
        <span class="pt-2">
          Policy
        </span>
        <span class="col-span-3">
          <Input type="text" />
        </span>
      </div>
    </span>
    <span slot="footer" class="w-full flex justify-between items-center">
      <button on:click={() => {modalOpen = true;}} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-cyan-400 shadow rounded-lg hover:bg-cyan-500 focus:ring-4 focus:outline-none focus:ring-blue-300 dark:bg-cyan-700 dark:hover:bg-cyan-800 dark:focus:ring-blue-700">
        <span class="text-sm text-light uppercase">
          Reset
        </span>
      </button>
      <button on:click={() => {modalOpen = true;}} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
        <span class="text-sm text-light uppercase">
          Create
        </span>
      </button>
    </span>
  </Modal>
</main>