<script lang="ts">
  import type { ApiKey } from "../../../api/authorities";

  export let apiKey: ApiKey;
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-2 place-items-start text-white">
    <span>
      {apiKey.id}
    </span>
    <span class="place-self-end">
      <i class="fa fa-trash cursor-pointer hover:text-zinc-500"></i>
    </span>
  </div>
</div>