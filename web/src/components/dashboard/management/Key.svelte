<script lang="ts">
  import type { Key } from "../../../api/authorities";
  import Button from "../../Button.svelte";
  import Modal from "../../Modal.svelte";

  export let authorityKey: Key;

  let asciiArmorModalEnabled: boolean = false;
  let trustSignatureModalEnabled: boolean = false;

  const showAsciiArmorModal = () => {
    asciiArmorModalEnabled = true;
  };

  const hideAsciiArmorModal = () => {
    asciiArmorModalEnabled = false;
  };

  const showTrustSignatureModal = () => {
    trustSignatureModalEnabled = true;
  };

  const hideTrustSignatureModal = () => {
    trustSignatureModalEnabled = false;
  };
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-4 place-items-start">
    <span class="break-all">
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
      <Button>
        <i class="fa fa-pen-to-square"></i>
      </Button>
      <Button>
        <i class="fa fa-trash"></i>
      </Button>
    </span>
  </div>
</div>

<Modal title={`ASCII Armor of Key ID ${authorityKey.keyId}`} enabled={asciiArmorModalEnabled} onClose={hideAsciiArmorModal}>
  <span slot="body">
    <pre class="text-xs">{authorityKey.asciiArmor || "<empty>"}</pre>
  </span>
</Modal>

<Modal title={`Trust Signature of Key ID ${authorityKey.keyId}`} enabled={trustSignatureModalEnabled} onClose={hideTrustSignatureModal}>
  <span slot="body">
    <pre class="text-xs">{authorityKey.trustSignature || "<empty>"}</pre>
  </span>
</Modal>