<script lang="ts">
  import TransparentButton from './TransparentButton.svelte';
  import CaretButton from './CaretButton.svelte';
  import Icon from './Icon.svelte';

  import type { Authority } from '@/api/authorities';

  export let authority: Authority;
  export let rulesShown: boolean = false;
  export let onAddRule: () => void = () => {};
  export let onToggleRules: () => void = () => {};
</script>

{#if authority.upstreamHostname}
  <span class="flex flex-wrap items-center gap-2 text-sm">
    <span class="break-all">
      {authority.upstreamHostname}/{authority.upstreamNamespace}
    </span>
    {#if authority.upstreamEnabled}
      <span
        class="px-2 rounded-lg text-xs uppercase bg-teal-200 dark:bg-teal-900">
        pulls through
      </span>
    {/if}
    <span class="flex items-center">
      <TransparentButton onClick={onAddRule} label="Add upstream rule">
        <Icon name="plus" />
      </TransparentButton>
      <span class="ml-1">{authority.rules?.length ?? 0} rules</span>
      {#if authority.rules?.length > 0}
        <CaretButton
          class="ml-1"
          onClick={onToggleRules}
          enabled={rulesShown}
          label="Show upstream rules" />
      {/if}
    </span>
  </span>
{/if}
