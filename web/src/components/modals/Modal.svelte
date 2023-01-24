<script lang="ts">
  import { clickOutside } from 'svelte-use-click-outside';

  import Button from '../inputs/Button.svelte';
  import KeyboardAction from '../inputs/KeyboardAction.svelte';
  import Icon from '../icons/Icon.svelte';

  export let title: string;
  export let enabled: boolean = false;
  export let onClose: () => void = () => {};

  const close = () => {
    enabled = false;
    onClose();
  };
</script>

{#if enabled}
  <KeyboardAction trigger={'Escape'} action={close} />

  <div 
    tabindex="-1"
    aria-hidden={!enabled}
    class="fixed top-0 left-0 right-0 z-50 w-screen h-screen bg-zinc-800/75 p-4 overflow-x-hidden overflow-y-auto md:inset-0 h-modal md:h-full"
  >
    <div class="relative w-full h-full flex justify-center items-center">
      <div use:clickOutside={close} class="relative w-full max-w-2xl bg-slate-50 rounded-lg shadow dark:bg-slate-900">
        <div class="flex items-start justify-between p-4 border-b rounded-t bg-teal-400 dark:bg-teal-700 dark:border-slate-600 dark:text-white">
          <h3 class="text-xl font-semibold text-slate-600 dark:text-slate-200">
            {title}
          </h3>
          <Button onClick={close}>
            <Icon name="close" width="1.25rem" height="1.25rem" />
          </Button>
        </div>
        <div class="p-6 max-h-96 text-slate-600 dark:text-slate-200 overflow-auto">
          <slot name="body"></slot>
        </div>
        {#if $$slots.footer}
          <div class="flex items-center text-xs text-slate-600 dark:text-slate-200 p-3 space-x-2 border-t border-slate-200 rounded-b dark:border-slate-600">
            <slot name="footer"></slot>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}