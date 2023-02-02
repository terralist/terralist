<script lang="ts">
  import { onMount } from "svelte";

  import type { InputType } from "../../lib/form";

  export let id: string = Math.random().toString();
  export let type: InputType = 'text';
  export let placeholder: string = "";
  export let value: string | string[] = "";
  export let disabled: boolean = false;
  export let slotPosition: "start" | "end" = "start";

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
    "disabled:opacity-50",
    "disabled:bg-slate-300",
    "dark:disabled:bg-slate-800",
    "dark:disabled:text-slate-300",
    $$slots?.default ? (
      slotPosition === "start" ? 'pl-8 pr-2' : 'pl-2 pr-8'
    ) : 'px-2',
  ];

  let ref: HTMLInputElement | HTMLTextAreaElement;
  let highlightClassList: string[];

  export function highlight(type: "none" | "success" | "error" = "none") {
    if (highlightClassList) {
      highlightClassList.forEach(className => ref.classList.remove(className));
    }

    highlightClassList = {
      "success": ["ring-1", "outline-none", "ring-green-500", "dark:ring-green-200"],
      "error":   ["ring-1", "outline-none", "ring-red-500", "dark:ring-red-200"],
      "none":    [],
    }[type];

    highlightClassList.forEach(className => ref.classList.add(className));
  }

  export function focus() {
    ref.focus();
  };

  export function blur() {
    ref.blur();
  }

  export function setValue(value?: string) {
    if (!value) {
      value = "";
    }

    value = value;
    ref.value = value;
  }

  const handleChange = () => {
    value = ref.value;
    onInput();
  }

  onMount(() => {
    if (ref && !["textarea"].includes(type)) {
      (ref as HTMLInputElement).type = type;
    }
  });
</script>

<svelte:options accessors={true}/>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div class={$$props.class} on:click={onClick}>
  {#if slotPosition === "start" && $$slots?.default}
    <slot></slot>
  {/if}
  {#if type === "textarea"}
  <textarea
    id={id}
    class="{classList.join(" ")} py-3 h-20"
    placeholder={placeholder}
    disabled={disabled}
    value={value}
    on:input={handleChange}
    bind:this={ref}
  />
  {:else}
  <input
    id={id}
    class="{classList.join(" ")} h-10"
    placeholder={placeholder}
    disabled={disabled}
    value={value}
    on:input={handleChange}
    bind:this={ref}
  />
  {/if}
  {#if slotPosition === "end" && $$slots?.default}
    <slot></slot>
  {/if}
</div>