<script lang="ts">
  import type { Key } from "../../../api/authorities";

  export let authorityKey: Key;
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-4 place-items-start text-white">
    <span>
      {authorityKey.keyId}
    </span>
    <span>
      <i class="fa fa-eye cursor-pointer hover:text-zinc-500"></i>
      <!-- {authorityKey.asciiArmor} -->
    </span>
    <span>
      <i class="fa fa-eye cursor-pointer hover:text-zinc-500"></i>
      <!-- {authorityKey.trustSignature} -->
    </span>
    <span class="place-self-end">
      <i class="fa fa-pen-to-square cursor-pointer hover:text-zinc-500"></i>
      <i class="fa fa-trash cursor-pointer hover:text-zinc-500"></i>
    </span>
  </div>
</div>