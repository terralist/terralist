<script lang="ts">
  import { useToggle } from '@/lib/hooks';

  import Icon from "./Icon.svelte";

  export let label: string;
  export let options: string[] = [];
  export let onSelect: (option: string) => void = () => {};

  const [open, toggle] = useToggle(false);

  const select = (option: string) => {
    toggle();
    onSelect(option);
  }
</script>

<button 
  on:click={toggle}
  class="
    text-slate-600
    dark:text-slate-200
    bg-teal-400
    hover:bg-teal-500
    dark:bg-teal-700
    dark:hover:bg-teal-800
    focus:ring-4
    focus:outline-none
    focus:ring-green-300
    dark:focus:ring-green-700
    font-medium
    rounded-lg
    text-sm
    px-4
    py-2.5
    w-40
    inline-flex
    justify-between
    gap-4
    items-center
    {$$props.class}
  " 
  type="button"
>
  {label}
  <Icon name="arrow-down" />
</button>

{#if $open}
  <div 
    class="
      z-10
      mt-2
      absolute
      block
      bg-white
      divide-y
      divide-gray-100
      rounded-lg
      shadow
      w-40
      max-w-40
      max-h-80
      overflow-y-auto
      dark:bg-gray-700
      {$$props.class}
    "
  >
    <ul class="py-2 text-sm text-gray-700 dark:text-gray-200">
      {#each options as option}
        <li>
          <button
            class="
              block
              px-4
              py-2
              w-full
              hover:bg-gray-100
              dark:hover:bg-gray-600
              dark:hover:text-white
              flex
              justify-start
              items-center
            "
            on:click={() => select(option)}
          >
            {option}
          </button>
        </li>
      {/each}
    </ul>
  </div>
{/if}