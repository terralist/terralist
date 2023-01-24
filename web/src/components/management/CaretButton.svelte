<script lang="ts">
  import Button from "../inputs/Button.svelte";

  export let className: string = "";
  export let onClick: () => void = () => {};
  export let enabled: boolean = false;
</script>

{#key enabled}
  <Button className={className} onClick={onClick}>
    <i class="fa fa-caret-{enabled ? 'up' : 'down'}"></i>
  </Button>
{/key}
