<script lang="ts">
  import TransparentButton from './TransparentButton.svelte';
  import Icon from './Icon.svelte';
  import ConfirmationModal from './ConfirmationModal.svelte';

  import type { Rule } from '@/api/rules';

  import { useFlag } from '@/lib/hooks';

  export let rule: Rule;
  export let authorityName: string;
  export let onDelete: (id: string) => void = () => {};

  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);

  const remove = () => {
    onDelete(rule.id);
  };
</script>

<div
  class="mt-2 mx-4"
  data-testid={`rule-${rule.kind}-${rule.name}-${rule.version}`}>
  <div
    class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-5 place-items-start">
    <span>{rule.kind}</span>
    <span class="break-all">{rule.name}</span>
    <span class="break-all">{rule.version}</span>
    <span>{rule.effect}</span>
    <span class="place-self-end">
      <TransparentButton onClick={showDeleteModal} label="Remove rule">
        <Icon name="trash" />
      </TransparentButton>
    </span>
  </div>
</div>

<ConfirmationModal
  title={`Remove a rule of ${authorityName}`}
  enabled={$deleteModalEnabled}
  onClose={hideDeleteModal}
  onSubmit={remove}>
  The <b>{rule.effect}</b> rule for {rule.kind} <b>{rule.name}</b>, versions
  <b>{rule.version}</b>, will no longer apply to the upstream of
  <b>{authorityName}</b>.
  <br /><br />
  Are you sure?
</ConfirmationModal>
