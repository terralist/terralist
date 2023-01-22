<script lang="ts">
  import Key from "./Key.svelte";
  import ApiKey from "./ApiKey.svelte";
  import type { Authority } from "../../../api/authorities";
  import { URLValidation } from "../../../lib/validation";
  import CaretButton from "./CaretButton.svelte";
  import Button from "../../Button.svelte";
  import ConfirmationModal from "../../ConfirmationModal.svelte";
  import { useFlag, useToggle } from "../../../api/hooks";
  import FormModal from "../../FormModal.svelte";

  export let authority: Authority;

  const [createKeyModalEnabled, showCreateKeyModal, hideCreateKeyModal] = useFlag(false);
  const [createApiKeyModalEnabled, showCreateApiKeyModal, hideCreateApiKeyModal] = useFlag(false);
  const [updateModalEnabled, showUpdateModal, hideUpdateModal] = useFlag(false);
  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);

  // To avoid a circular dependency, we will create a wrapper fn for toggleShowKeys
  const [showKeys, _toggleShowKeys] = useToggle(false);
  const [showApiKeys, toggleShowApiKeys] = useToggle(false, () => {
    if ($showKeys) {
      _toggleShowKeys();
    }
  });

  const toggleShowKeys = () => {
    if ($showApiKeys) {
      toggleShowApiKeys();
    }

    _toggleShowKeys();
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
      <Button onClick={showCreateKeyModal}>
        <i class="fa fa-plus w-3 h-3"></i>
      </Button>
      <span class="ml-0 md:ml-2">
        {authority.keys.length}
      </span>
      {#if authority.keys.length > 0}
        <CaretButton className="ml-2" onClick={toggleShowKeys} enabled={$showKeys} />
      {/if}
    </span>
    <span class="flex flex-col md:flex-row justify-center items-center">
      <Button onClick={showCreateApiKeyModal}>
        <i class="fa fa-plus w-3 h-3"></i>
      </Button>
      <span class="ml-0 md:ml-2">
        {authority.apiKeys.length}
      </span>
      {#if authority.apiKeys.length > 0}
        <CaretButton className="ml-0 md:ml-2" onClick={toggleShowApiKeys} enabled={$showApiKeys} />
      {/if}
    </span>
    <span class="place-self-end flex justify-center items-center">
      <Button onClick={showUpdateModal}>
        <i class="fa fa-pen-to-square"></i>
      </Button>
      <Button onClick={showDeleteModal}>
        <i class="fa fa-trash"></i>
      </Button>
    </span>
  </div>
  {#if $showKeys}
    <div class="w-full p-2 px-6 grid grid-cols-4 lg:grid-cols-8 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span class="lg:col-span-5">
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
        <Key authorityKey={key} authorityName={authority.name} isAlone={authority.keys.length === 1} />
      {/key}
    {/each}
  {/if}
  {#if $showApiKeys}
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
        <ApiKey apiKey={apiKey} authorityName={authority.name} />
      {/key}
    {/each}
  {/if}

  <FormModal 
    title={`Update authority ${authority.name}`}
    enabled={$updateModalEnabled}
    onClose={hideUpdateModal}
    onSubmit={() => {}}

    entries={[
      {
        name: "Name",
        type: "text",
        disabled: true,
        value: authority.name,
      },
      {
        name: "Policy",
        type: "text",
        value: authority.policyUrl,
        validations: [URLValidation()],
      },
    ]}
  />

  <ConfirmationModal 
    title={`Remove authority ${authority.name}`} 
    enabled={$deleteModalEnabled}
    onClose={hideDeleteModal}
  >
    Removing the <b>{authority.name}</b> authority will also remove all artifacts uploaded to the <b>{authority.name}</b> namespace.
    <br/><br/>
    Are you sure?
  </ConfirmationModal>

  <FormModal 
    title={`Add a new authority key to ${authority.name}`}
    enabled={$createKeyModalEnabled}
    onClose={hideCreateKeyModal}
    onSubmit={() => {}}

    entries={[
      {
        name: "Key ID",
        required: true,
        type: "text",
        validations: [],
      },
      {
        name: "ASCII Armor",
        type: "textarea",
        validations: [],
      },
      {
        name: "Trust Signature",
        type: "textarea",
        validations: [],
      },
    ]}
  />

  <ConfirmationModal 
    title={`Add a new API key to ${authority.name}`} 
    enabled={$createApiKeyModalEnabled}
    onClose={hideCreateApiKeyModal}
  >
    Are you sure?
  </ConfirmationModal>
</div>