<script lang="ts">
  import Icon from "./Icon.svelte";
  import TransparentButton from "./TransparentButton.svelte";

  export let onClick: () => void = () => {};
  export let enabled: boolean = false;
</script>

{#key enabled}
  <TransparentButton class={$$props.class} onClick={onClick}>
    <Icon name={enabled ? "arrow-up" : "arrow-down"} />
  </TransparentButton>
{/key}
