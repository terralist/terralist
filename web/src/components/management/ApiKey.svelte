<script lang="ts">
  import Button from "../inputs/Button.svelte";

  import Modal from "../modals/Modal.svelte";
  import ConfirmationModal from "../modals/ConfirmationModal.svelte";
  import ErrorModal from "../modals/ErrorModal.svelte";

  import type { ApiKey } from "../../api/authorities";

  import { useFlag } from "../../lib/hooks";


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
      <Button onClick={showApiKeyModal}>
        <i class="fa fa-eye"></i>
      </Button>
    </span>
    <span class="place-self-end">
      <Button onClick={showDeleteModal}>
        <i class="fa fa-trash"></i>
      </Button>
    </span>
  </div>
</div>

<Modal title="View API Key" enabled={$apiKeyModalEnabled} onClose={hideApiKeyModal}>
  <span slot="body">
    <div class="flex justify-between items-center">
      <pre class="text-xs">{apiKey.id}</pre>
      {#key $clipboardUpdated}
        <Button onClick={updateClipboard} disabled={$clipboardUpdated}>
          <i class="fa {$clipboardUpdated ? 'fa-clipboard-check' : 'fa-clipboard'}"></i>
        </Button>
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