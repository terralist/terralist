<script lang="ts">
  import Modal from './Modal.svelte'

  export let title: string;
  export let enabled: boolean = false;
  export let onClose: () => void = () => { };
  export let onConfirm: () => void = () => { };

  const close = () => {
    enabled = false;
    onClose();
  };

  const confirm = () => {
    close();
    onConfirm();
  }
</script>

<Modal title={title} enabled={enabled} onClose={close}>
  <span slot="body">
    <slot></slot>
  </span>
  
  <span slot="footer" class="w-full flex justify-end items-center gap-4">
    <button on:click={close} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-cyan-400 shadow rounded-lg hover:bg-cyan-500 focus:ring-4 focus:outline-none focus:ring-blue-300 dark:bg-cyan-700 dark:hover:bg-cyan-800 dark:focus:ring-blue-700">
      <span class="text-sm text-light uppercase">
        Cancel
      </span>
    </button>
    <button on:click={confirm} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      <span class="text-sm text-light uppercase">
        Continue
      </span>
    </button>
  </span>
</Modal>
