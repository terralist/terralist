<script lang="ts">
  import Icon from "../icons/Icon.svelte";
  
  export let brandName: String;

  let icon: string = (() => {
    switch (brandName.toLowerCase()) {
      // Any icons overrides should be written here
      // e.g. case "microsoft": return "office365";
      default: return brandName.toLowerCase();
    }
  })();
</script>

<button
  class="
    h-14
    outline-0
    cursor-pointer
    w-full
    bg-teal-400
    text-slate-100
    border-0
    rounded-xl
    text-base
    text-semibold
    tracking-wide
    grid
    grid-cols-6
    place-items-center
  "
>
  <Icon name={icon} />
  <span class="col-span-5">Continue with {brandName}</span>
</button>