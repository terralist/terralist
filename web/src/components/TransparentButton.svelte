<script lang="ts">
  export let onClick: () => void = () => {};
  export let disabled: boolean = false;
</script>

<button
  type="button"
  on:click={onClick}
  class="
    text-slate-600
    dark:text-slate-200
    bg-transparent
    hover:bg-teal-200
    dark:hover:bg-teal-900
    hover:text-slate-900
    dark:hover:text-slate-50
    rounded-lg
    text-sm
    p-1.5
    ml-auto
    inline-flex
    items-center
    disabled:opacity-75
    disabled:bg-transparent
    disabled:hover:bg-transparent
    disabled:hover:dark:bg-transparent
    {$$props.class}
  "
  disabled={disabled}>
  <slot></slot>
</button>