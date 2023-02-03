<script lang="ts">
  import TransparentButton from "./TransparentButton.svelte";
  import CaretButton from "./CaretButton.svelte";
  import Icon from "./Icon.svelte";

  import FormModal from "./FormModal.svelte";
  import ConfirmationModal from "./ConfirmationModal.svelte";
  import ErrorModal from "./ErrorModal.svelte";

  import Key from "./Key.svelte";
  import ApiKey from "./ApiKey.svelte";

  import { createApiKey, createKey, deleteApiKey, deleteKey, type Authority } from "@/api/authorities";

  import { URLValidation } from "@/lib/validation";
  import { useFlag, useToggle } from "@/lib/hooks";

  export let authority: Authority;
  export let onUpdate: (id: string, authority: Authority) => void = () => {};
  export let onDelete: (id: string) => void = () => {};

  let errorMessage: string = "";

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

  const update = (entries: Map<string, any>) => {
    onUpdate(
      authority.id, 
      {...authority, policyUrl: entries.get("policyUrl")}
    );
  };

  const remove = () => {
    onDelete(authority.id);
  };

  const createKeySubmit = (entries: Map<string, any>) => {
    let result = createKey(authority, entries.get("key_id"), entries.get("ascii_armor"), entries.get("trust_signature"));

    if (result.status === 'OK') {
      authority.keys = [...authority.keys, result.data];
    } else {
      errorMessage = result.message;
    }
  };

  const onKeyDelete = (id: string) => {
    let result = deleteKey(authority, authority.keys.find(k => k.id === id));

    if (result.status === 'OK') {
      authority.keys = [...authority.keys.filter(k => k.id !== id)];
    } else {
      errorMessage = result.message;
    }
  };

  const createApiKeySubmit = () => {
    let result = createApiKey(authority);

    if (result.status === 'OK') {
      authority.apiKeys = [...authority.apiKeys, result.data];
    } else {
      errorMessage = result.message;
    }
  };

  const onApiKeyDelete = (id: string) => {
    let result = deleteApiKey(authority, authority.apiKeys.find(ak => ak.id === id));

    if (result.status === 'OK') {
      authority.apiKeys = [...authority.apiKeys.filter(ak => ak.id !== id)];
    } else {
      errorMessage = result.message;
    }
  };
</script>

<div class="mb-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-6 lg:grid-cols-10 place-items-start">
    <span class="col-span-2 lg:col-span-6">{authority.name}</span>
    <span>
      {#if authority.policyUrl}
        <a href={authority.policyUrl} target="_blank" rel="noreferrer">
          <TransparentButton>
            <Icon name="external-link" />
          </TransparentButton>
        </a>
      {:else}
        <span>-</span>
      {/if}
    </span>
    <span class="flex flex-col md:flex-row justify-center items-center">
      <TransparentButton onClick={showCreateKeyModal}>
        <Icon name="plus" />
      </TransparentButton>
      <span class="ml-0 md:ml-2">
        {authority.keys.length}
      </span>
      {#if authority.keys.length > 0}
        <CaretButton class="ml-0 md:ml-2" onClick={toggleShowKeys} enabled={$showKeys} />
      {/if}
    </span>
    <span class="flex flex-col md:flex-row justify-center items-center">
      <TransparentButton onClick={showCreateApiKeyModal}>
        <Icon name="plus" />
      </TransparentButton>
      <span class="ml-0 md:ml-2">
        {authority.apiKeys.length}
      </span>
      {#if authority.apiKeys.length > 0}
        <CaretButton class="ml-0 md:ml-2" onClick={toggleShowApiKeys} enabled={$showApiKeys} />
      {/if}
    </span>
    <span class="place-self-end flex justify-center items-center">
      <TransparentButton onClick={showUpdateModal}>
        <Icon name="edit-box" />
      </TransparentButton>
      <TransparentButton onClick={showDeleteModal}>
        <Icon name="trash" />
      </TransparentButton>
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
        <Key authorityKey={key} authorityName={authority.name} isAlone={authority.keys.length === 1} onDelete={onKeyDelete} />
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
        <ApiKey apiKey={apiKey} authorityName={authority.name} onDelete={onApiKeyDelete} />
      {/key}
    {/each}
  {/if}

  <FormModal 
    title={`Update authority ${authority.name}`}
    enabled={$updateModalEnabled}
    onClose={hideUpdateModal}
    onSubmit={update}

    entries={[
      {
        id: "name",
        name: "Name",
        type: "text",
        disabled: true,
        value: authority.name,
      },
      {
        id: "policyUrl",
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
    onSubmit={remove}
  >
    Removing the <b>{authority.name}</b> authority will also remove all artifacts uploaded to the <b>{authority.name}</b> namespace.
    <br/><br/>
    Are you sure?
  </ConfirmationModal>

  <FormModal 
    title={`Add a new authority key to ${authority.name}`}
    enabled={$createKeyModalEnabled}
    onClose={hideCreateKeyModal}
    onSubmit={createKeySubmit}

    entries={[
      {
        id: "key_id",
        name: "Key ID",
        required: true,
        type: "text",
        validations: [],
      },
      {
        id: "ascii_armor",
        name: "ASCII Armor",
        type: "textarea",
        validations: [],
      },
      {
        id: "trust_signature",
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
    onSubmit={createApiKeySubmit}
  >
    Are you sure?
  </ConfirmationModal>

  {#if errorMessage}
    <ErrorModal bind:message={errorMessage} />
  {/if}
</div>