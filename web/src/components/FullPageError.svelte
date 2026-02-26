<script lang="ts">
  import Icon from '@/components/Icon.svelte';

  export let code: number = 404;
  export let message: string =
    "The resource you're looking for was not found on the server.";

  const codeDigits =
    code >= 100 && code < 1000 ? code.toString().split('') : [];
</script>

<div
  class="absolute top-0 left-0 flex justify-center items-center text-center w-screen h-screen -z-10">
  <div class="flex flex-col justify-center items-center gap-4">
    <span class="flex">
      {#each codeDigits as digit, index (`${index}-${digit}`)}
        <Icon
          name="number-{digit}"
          width="32px"
          height="32px"
          class="fill-zinc-900 dark:fill-zinc-100" />
      {/each}
    </span>
    <p class="font-medium text-medium text-zinc-900 dark:text-zinc-100">
      {message}
    </p>
  </div>
</div>
