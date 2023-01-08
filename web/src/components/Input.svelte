<script lang="ts">
  import { onMount } from "svelte";

  import type { InputType } from "../lib/form";

  export let id: string = Math.random().toString();
  export let className: string = "";
  export let type: InputType = 'text';
  export let placeholder: string = "";
  export let value: string = "";

  export let onClick: () => void = () => {};
  export let onInput: () => void = () => {};

  const classList: string[] = [
    "w-full",
    "pr-2",
    "bg-slate-100",
    "dark:bg-slate-800",
    "text-slate-800",
    "dark:text-slate-200",
    "shadow",
    "border-none",
    "text-sm",
    "rounded-lg",
    "focus:ring-0",
    "outline-none",
    $$slots?.default ? 'pl-8' : 'pl-2',
  ];

  let ref: HTMLInputElement | HTMLTextAreaElement;
  let currentHighlight: string[];

  export function highlight(type: "none" | "success" | "error" = "none") {
    if (currentHighlight) {
      for (let className of currentHighlight) {
        ref.classList.remove(className);
      }
    }

    currentHighlight = (() => {
      switch (type) {
        case "success": return ["ring-1", "outline-none", "ring-green-500", "dark:ring-green-200"];
        case "error": return ["ring-1", "outline-none", "ring-red-500", "dark:ring-red-200"];
        case "none": return [];
      }
    })();

    for (let className of currentHighlight) {
      ref.classList.add(className);
    }
  }

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
    if (ref && type !== "textarea") {
      ref.type = type;
    }
  });
</script>

<svelte:options accessors={true}/>

<div class={className}>
  {#if $$slots?.default}
    <slot></slot>
  {/if}
  {#if type === "textarea"}
  <textarea
    id={id}
    class="{classList.join(" ")} py-3 h-20"
    placeholder={placeholder}
    on:click={onClick}
    on:input={handleChange}
    bind:this={ref}
  />
  {:else}
  <input
    id={id}
    class="{classList.join(" ")} h-10"
    placeholder={placeholder}
    on:click={onClick}
    on:input={handleChange}
    bind:this={ref}
  />
  {/if}
</div>