<script lang="ts">
  import type { ApiKey } from "../../../api/authorities";
  import Button from "../../Button.svelte";
  import ConfirmationModal from "../../ConfirmationModal.svelte";

  import { useFlag } from "../../../api/hooks";

  export let apiKey: ApiKey;
  export let authorityName: string;

  let [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-2 place-items-start">
    <span>
      {apiKey.id}
    </span>
    <span class="place-self-end">
      <Button onClick={showDeleteModal}>
        <i class="fa fa-trash"></i>
      </Button>
    </span>
  </div>
</div>

<ConfirmationModal
  title={`Remove Api Key ${apiKey.id} of ${authorityName}`} 
  enabled={$deleteModalEnabled}
  onClose={hideDeleteModal}
>
  Are you sure?
</ConfirmationModal>