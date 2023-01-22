<script lang="ts">
  import type { Key } from "../../../api/authorities";
  import { useFlag } from "../../../api/hooks";
  import Button from "../../Button.svelte";
  import ConfirmationModal from "../../ConfirmationModal.svelte";
  import Modal from "../../Modal.svelte";

  export let authorityKey: Key;
  export let authorityName: string;
  export let isAlone: boolean = false;

  const [asciiArmorModalEnabled, showAsciiArmorModal, hideAsciiArmorModal] = useFlag(false);
  const [trustSignatureModalEnabled, showTrustSignatureModal, hideTrustSignatureModal] = useFlag(false);
  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-4 lg:grid-cols-8 place-items-start">
    <span class="break-all lg:col-span-5">
      {authorityKey.keyId}
    </span>
    <span>
      <Button onClick={showAsciiArmorModal}>
        <i class="fa fa-eye"></i>
      </Button>
    </span>
    <span>
      <Button onClick={showTrustSignatureModal}>
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

<Modal title={`ASCII Armor of Key ID ${authorityKey.keyId}`} enabled={$asciiArmorModalEnabled} onClose={hideAsciiArmorModal}>
  <span slot="body">
    <pre class="text-xs">{authorityKey.asciiArmor || "<empty>"}</pre>
  </span>
</Modal>

<Modal title={`Trust Signature of Key ID ${authorityKey.keyId}`} enabled={$trustSignatureModalEnabled} onClose={hideTrustSignatureModal}>
  <span slot="body">
    <pre class="text-xs">{authorityKey.trustSignature || "<empty>"}</pre>
  </span>
</Modal>

<ConfirmationModal 
  title={`Remove Key ID ${authorityKey.keyId} of ${authorityName}`} 
  enabled={$deleteModalEnabled}
  onClose={hideDeleteModal}
>
  {#if isAlone}
    This is the last key of <b>{authorityName}</b> authority. 
    Removing it will also remove the authority and all artifacts uploaded to the <b>{authorityName}</b> namespace.
    <br/><br/>
  {/if}
  Are you sure?
</ConfirmationModal>
