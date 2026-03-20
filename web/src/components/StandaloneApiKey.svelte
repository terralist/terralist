<script lang="ts">
  import TransparentButton from './TransparentButton.svelte';
  import Icon from './Icon.svelte';

  import Modal from './Modal.svelte';
  import ConfirmationModal from './ConfirmationModal.svelte';
  import ErrorModal from './ErrorModal.svelte';

  import type { StandaloneApiKey } from '@/api/standaloneApiKeys';

  import { useFlag } from '@/lib/hooks';

  export let apiKey: StandaloneApiKey;
  export let onDelete: (id: string) => void = () => {};

  const [clipboardUpdated, setClipboardUpdated, resetClipboardUpdated] =
    useFlag(false);

  const [viewModalEnabled, showViewModal, hideViewModal] = useFlag(false);
  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);

  let errorMessage: string = '';

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

  const formatEffect = (effect: string) => {
    return effect === 'allow' ? '✓' : '✗';
  };
</script>

<div class="mt-2">
  <div
    class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-4 place-items-start items-center">
    <span class="truncate">{apiKey.name}</span>
    <span class="text-xs truncate">{apiKey.scope}</span>
    <span class="text-xs">
      {apiKey.policies.length}
      {apiKey.policies.length === 1 ? 'policy' : 'policies'}
    </span>
    <span class="place-self-end flex gap-1">
      <TransparentButton onClick={showViewModal}>
        <Icon name="eye" />
      </TransparentButton>
      <TransparentButton onClick={showDeleteModal}>
        <Icon name="trash" />
      </TransparentButton>
    </span>
  </div>
</div>

<Modal
  title="API Key: {apiKey.name}"
  enabled={$viewModalEnabled}
  onClose={hideViewModal}>
  <span slot="body">
    <div class="space-y-4">
      <div>
        <p class="text-xs uppercase text-zinc-400 mb-1">Key</p>
        <div
          class="flex justify-between items-center bg-slate-100 dark:bg-slate-800 rounded-lg p-2">
          <pre class="text-xs">{apiKey.id}</pre>
          {#key $clipboardUpdated}
            <TransparentButton
              onClick={updateClipboard}
              disabled={$clipboardUpdated}>
              <Icon name={$clipboardUpdated ? 'check' : 'clipboard'} />
            </TransparentButton>
          {/key}
        </div>
      </div>

      <div>
        <p class="text-xs uppercase text-zinc-400 mb-1">Scope</p>
        <p class="text-sm">{apiKey.scope}</p>
      </div>

      <div>
        <p class="text-xs uppercase text-zinc-400 mb-1">Created by</p>
        <p class="text-sm">{apiKey.createdBy}</p>
      </div>

      {#if apiKey.expiration}
        <div>
          <p class="text-xs uppercase text-zinc-400 mb-1">Expires</p>
          <p class="text-sm">{apiKey.expiration}</p>
        </div>
      {/if}

      <div>
        <p class="text-xs uppercase text-zinc-400 mb-1">Policies</p>
        <div class="text-xs">
          <div
            class="grid grid-cols-5 gap-2 font-semibold uppercase text-zinc-400 mb-1">
            <span>Resource</span>
            <span>Action</span>
            <span class="col-span-2">Object</span>
            <span>Effect</span>
          </div>
          {#each apiKey.policies as policy (policy.id)}
            <div
              class="grid grid-cols-5 gap-2 py-1 border-t border-slate-200 dark:border-slate-600">
              <span>{policy.resource}</span>
              <span>{policy.action}</span>
              <span class="col-span-2 truncate">{policy.object}</span>
              <span>{formatEffect(policy.effect)}</span>
            </div>
          {/each}
        </div>
      </div>
    </div>
  </span>
</Modal>

<ConfirmationModal
  title={`Remove API Key ${censor(apiKey.id)}`}
  enabled={$deleteModalEnabled}
  onClose={hideDeleteModal}
  onSubmit={remove}>
  Are you sure you want to delete the API key <strong>{apiKey.name}</strong>?
</ConfirmationModal>

{#if errorMessage}
  <ErrorModal bind:message={errorMessage} />
{/if}
