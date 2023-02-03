<script lang="ts">
  export let onClick: () => void = () => {};
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
    hover:bg-teal-500
    focus:ring-4
    focus:outline-none
    focus:ring-green-300
    dark:bg-teal-700
    dark:hover:bg-teal-800
    dark:focus:ring-green-700
    {$$props.class}
  "
  on:click={onClick}
>
  <slot></slot>
</button>