<script lang="ts">
  import Icon from './Icon.svelte';
  import TransparentButton from './TransparentButton.svelte';

  export let onClick: () => void = () => {};
  export let enabled: boolean = false;
  export let label: string | undefined = undefined;
</script>

{#key enabled}
  <TransparentButton class={$$props.class} {onClick} {label}>
    <Icon name={enabled ? 'arrow-up' : 'arrow-down'} />
  </TransparentButton>
{/key}
