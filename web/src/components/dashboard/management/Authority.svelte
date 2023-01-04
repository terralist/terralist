<script lang="ts">
  import Key from "./Key.svelte";
  import ApiKey from "./ApiKey.svelte";
  import type { Authority } from "../../../api/authorities";
  import CaretButton from "./CaretButton.svelte";
    import Button from "../../Button.svelte";

  export let authority: Authority;

  let showKeys: boolean = false;
  let showApiKeys: boolean = false;

  const toggleShowKeys = () => {
    showApiKeys = false;
    showKeys = !showKeys;
  };

  const toggleShowApiKeys = () => {
    showKeys = false;
    showApiKeys = !showApiKeys;
  }
</script>

<div class="mb-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-6 lg:grid-cols-10 place-items-start">
    <span class="col-span-2 lg:col-span-6">{authority.name}</span>
    <span>
      {#if authority.policyUrl}
        <a href={authority.policyUrl} target="_blank" rel="noreferrer">
          <Button>
            <i class="fa fa-arrow-up-right-from-square"></i>
          </Button>
        </a>
      {:else}
        <span>-</span>
      {/if}
    </span>
    <span class="flex flex-col md:flex-row justify-center items-center">
      <Button>
        <i class="fa fa-plus w-3 h-3"></i>
      </Button>
      <span class="ml-0 md:ml-2">
        {authority.keys.length}
      </span>
      {#if authority.keys.length > 0}
        <CaretButton className="ml-2" onClick={toggleShowKeys} enabled={showKeys} />
      {/if}
    </span>
    <span class="flex flex-col md:flex-row justify-center items-center">
      <Button>
        <i class="fa fa-plus w-3 h-3"></i>
      </Button>
      <span class="ml-0 md:ml-2">
        {authority.apiKeys.length}
      </span>
      {#if authority.apiKeys.length > 0}
        <CaretButton className="ml-0 md:ml-2" onClick={toggleShowApiKeys} enabled={showApiKeys} />
      {/if}
    </span>
    <span class="place-self-end flex justify-center items-center">
      <Button>
        <i class="fa fa-pen-to-square"></i>
      </Button>
      <Button>
        <i class="fa fa-trash"></i>
      </Button>
    </span>
  </div>
  {#if showKeys}
    <div class="w-full p-2 px-6 grid grid-cols-4 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span>
        Key ID
      </span>
      <span>
        ASCII Armor
      </span>
      <span>
        Trust Signature
      </span>
      <span class="place-self-end">
        Actions
      </span>
    </div>
    {#each authority.keys as key}
      {#key key.id}
        <Key authorityKey={key} />
      {/key}
    {/each}
  {/if}
  {#if showApiKeys}
    <div class="w-full p-2 px-6 grid grid-cols-2 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span>
        Api Key
      </span>
      <span class="place-self-end">
        Actions
      </span>
    </div>
    {#each authority.apiKeys as apiKey}
      {#key apiKey.id}
        <ApiKey apiKey={apiKey} />
      {/key}
    {/each}
  {/if}
</div>