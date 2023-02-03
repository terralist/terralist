<script lang="ts">
  import TransparentButton from "./TransparentButton.svelte";
  import Icon from "./Icon.svelte";

  import Modal from "./Modal.svelte";
  import ConfirmationModal from "./ConfirmationModal.svelte";
  import ErrorModal from "./ErrorModal.svelte";

  import type { ApiKey } from "@/api/authorities";

  import { useFlag } from "@/lib/hooks";


  export let apiKey: ApiKey;
  export let authorityName: string;
  export let onDelete: (id: string) => void = () => {};

  const [clipboardUpdated, setClipboardUpdated, resetClipboardUpdated] = useFlag(false);

  const [apiKeyModalEnabled, showApiKeyModal, hideApiKeyModal] = useFlag(false);
  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);

  let errorMessage: string = "";

  const censor = (value: string) => {
    return `****${value.slice(-4)}`;
  };

  const updateClipboard = () => {
    navigator.clipboard.writeText(apiKey.id);
    setClipboardUpdated();
    setTimeout(resetClipboardUpdated, 1000);
  };

  const remove = () => {
    onDelete(apiKey.id);
  };
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-2 place-items-start">
    <span>
      <TransparentButton onClick={showApiKeyModal}>
        <Icon name="eye" />
      </TransparentButton>
    </span>
    <span class="place-self-end">
      <TransparentButton onClick={showDeleteModal}>
        <Icon name="trash" />
      </TransparentButton>
    </span>
  </div>
</div>

<Modal title="View API Key" enabled={$apiKeyModalEnabled} onClose={hideApiKeyModal}>
  <span slot="body">
    <div class="flex justify-between items-center">
      <pre class="text-xs">{apiKey.id}</pre>
      {#key $clipboardUpdated}
        <TransparentButton onClick={updateClipboard} disabled={$clipboardUpdated}>
          <Icon name={$clipboardUpdated ? 'check' : 'clipboard'} />
        </TransparentButton>
      {/key}
    </div>
  </span>
</Modal>

<ConfirmationModal
  title={`Remove API Key ${censor(apiKey.id)} of ${authorityName}`} 
  enabled={$deleteModalEnabled}
  onClose={hideDeleteModal}
  onSubmit={remove}
>
  Are you sure?
</ConfirmationModal>

{#if errorMessage}
  <ErrorModal bind:message={errorMessage} />
{/if}