<script lang="ts">
  import Icon from "../icons/Icon.svelte";
import Button from "../inputs/Button.svelte";

  export let onClick: () => void = () => {};
  export let enabled: boolean = false;
</script>

{#key enabled}
  <Button class={$$props.class} onClick={onClick}>
    <Icon name={enabled ? "arrow-up" : "arrow-down"} />
  </Button>
{/key}
