<script lang="ts">
  export let className: string = "";
  export let onClick: (e) => void = () => {};
  export let enabled: boolean = false;
</script>

{#key enabled}
  <button class={className} on:click={onClick}>
    <i class="fa fa-caret-{enabled ? 'up' : 'down'} cursor-pointer hover:text-zinc-500"></i>
  </button>
{/key}
