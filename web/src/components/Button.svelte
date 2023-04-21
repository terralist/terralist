<script lang="ts">
  export let onClick: () => void = () => {};
  export let disabled: boolean | null = null;
</script>

<button
  class="
    w-full
    inline-flex
    justify-center
    items-center
    py-2
    px-3
    text-sm
    font-medium
    text-center
    text-slate-600
    dark:text-slate-200
    bg-teal-400
    rounded-lg
    enabled:hover:bg-teal-500
    enabled:focus:ring-4
    enabled:focus:outline-none
    enabled:focus:ring-green-300
    dark:bg-teal-700
    dark:enabled:hover:bg-teal-800
    dark:enabled:focus:ring-green-700
    disabled:opacity-75
    {$$props.class}
  "
  on:click={onClick}
  {disabled}
>
  <slot></slot>
</button>