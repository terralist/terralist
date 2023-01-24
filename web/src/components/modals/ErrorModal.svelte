<script lang="ts">
  import Icon from '../icons/Icon.svelte';
import Button from '../inputs/Button.svelte';

  import Modal from './Modal.svelte'

  export let message: string;

  let enabled: boolean = true;

  const close = () => {
    enabled = false;
    message = "";
  };
</script>

<Modal title={"An error occurred"} enabled={enabled} onClose={close}>
  <span slot="body" class="flex flex-col justify-center items-center">
    <Icon name="alert" width="48px" height="48px" class="fill-red-400 dark:fill-red-400" />

    <span class="justify-self-start">
      {message}
    </span>
  </span>
  
  <span slot="footer" class="w-full flex justify-end items-center gap-4">
    <Button onClick={close} class="text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      <span class="text-sm text-light uppercase">
        Continue
      </span>
    </Button>
  </span>
</Modal>