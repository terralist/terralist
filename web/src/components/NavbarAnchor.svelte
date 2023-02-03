<script lang="ts">
  import { link } from "svelte-spa-router";
  import active from "svelte-spa-router/active";

  import Icon from "./Icon.svelte";

  let show: boolean = false;

  export let title: string;
  export let tooltip: string = title;
  export let icon: string;
  export let href: string = "#";
</script>

<a
  href={href}
  use:link
  use:active={{className: "md:bg-slate-200/25 dark:md:bg-slate-800/25"}}
  class="bg-transparent rounded-lg p-2 flex items-center lg:block fill-slate-200 md:fill-slate-900 {$$props.class}"
>
  <span
    on:mouseover={() => {show = true}}
    on:mouseout={() => {show = false}}
    on:focus={() => {show = true}}
    on:blur={() => {show = false}}
    class="cursor-pointer"
  >
    <Icon 
      name={icon}
      width="1.25rem"
      height="1.25rem"
      class="fill-inherit pt-0.5 lg:pt-0"
    />
    {#if show}
      <div
        class="hidden lg:block text-sm absolute text-slate-50 dark:text-slate-800 bg-slate-800 dark:bg-slate-50 rounded-lg p-2 transform translate-y-2 -translate-x-1/2">
        {tooltip}
      </div>
    {/if}
  </span>
  <span class="block pl-2 lg:pl-0 lg:hidden">
    {title}
  </span>
</a>