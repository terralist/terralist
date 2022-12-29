<script lang="ts">
  let show: boolean = false;

  export let className: String = "";
  export let title: String;
  export let tooltip: String = title;
  export let icon: String;
  export let iconClass: "" | "solid" = "";
  export let clickHandler: () => void = () => {};
</script>

<button class="bg-transparent pb-2 flex items-center lg:block {className}" on:click={clickHandler}>
  <span
    on:mouseover={() => {show = true}}
    on:mouseout={() => {show = false}}
    on:focus={() => {show = true}}
    on:blur={() => {show = false}}
    class="cursor-pointer"
  >
    <i class="fa{iconClass ? `-${iconClass}` : ''} fa-{icon} pt-0.5 lg:pt-0 w-4 h-4"></i>
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
</button>