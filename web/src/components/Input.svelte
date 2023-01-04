<script lang="ts">
  import { onMount } from "svelte";

  export let className: string = "";
  export let type: 'email' | 'text' | 'password' | 'number' = 'text';
  export let placeholder: string = "";
  export let value: string = "";

  export let onClick: () => void = () => {};
  export let onInput: () => void = () => {};

  let ref: HTMLInputElement;

  export function focus() {
    ref.focus();
  };

  export function blur() {
    ref.blur();
  }

  const handleChange = () => {
    value = ref.value;
    onInput();
  }

  onMount(() => {
    if (ref) {
      ref.type = type;
    }
  });
</script>

<div class={className}>
  {#if $$slots?.default}
    <slot></slot>
  {/if}
  <input
    class="
      w-full
      h-10
      {$$slots?.default ? 'pl-8' : 'pl-2'}
      pr-2
      bg-slate-100
      dark:bg-slate-800
      text-slate-800
      dark:text-slate-200
      shadow
      border-none
      text-sm
      rounded-lg
      focus:ring-0
      outline-none
    "
    placeholder={placeholder}
    on:click={onClick}
    on:input={handleChange}
    bind:this={ref}
  />
</div>