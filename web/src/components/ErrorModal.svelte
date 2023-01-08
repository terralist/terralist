<script lang="ts">
  import Button from './Button.svelte';
  import Modal from './Modal.svelte'

  export let errorMessage: string;

  let enabled: boolean = true;

  const close = () => {
    enabled = false;
  };
</script>

<Modal title={"An error occurred"} enabled={enabled} onClose={close}>
  <span slot="body" class="flex flex-col justify-center items-center">
    <svg 
      height="32" 
      style="overflow:visible;enable-background:new 0 0 32 32" 
      viewBox="0 0 32 32" 
      width="32" 
      xml:space="preserve"
      xmlns="http://www.w3.org/2000/svg"
      xmlns:xlink="http://www.w3.org/1999/xlink"
    >
      <g>
        <g>
          <g>
            <circle cx="16" cy="16" id="BG" r="16" style="fill:#D72828;"/>
            <path d="M14.5,25h3v-3h-3V25z M14.5,6v13h3V6H14.5z" style="fill:#E6E6E6;"/>
          </g>
        </g>
      </g>
    </svg>
    <span class="justify-self-start">
      {errorMessage}
    </span>
  </span>
  
  <span slot="footer" class="w-full flex justify-end items-center gap-4">
    <Button onClick={close} className="text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      <span class="text-sm text-light uppercase ">
        Continue
      </span>
    </Button>
  </span>
</Modal>