<script lang="ts">
  import { clickOutside } from 'svelte-use-click-outside';

  import Button from '../inputs/Button.svelte';
  import KeyboardAction from '../inputs/KeyboardAction.svelte';

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
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
            </svg>
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