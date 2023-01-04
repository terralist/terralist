<script lang="ts">
  import type { ApiKey } from "../../../api/authorities";
  import Button from "../../Button.svelte";

  export let apiKey: ApiKey;
</script>

<div class="mt-2 mx-4">
  <div class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-2 place-items-start">
    <span>
      {apiKey.id}
    </span>
    <span class="place-self-end">
      <Button>
        <i class="fa fa-trash"></i>
      </Button>
    </span>
  </div>
</div>