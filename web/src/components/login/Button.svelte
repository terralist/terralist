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
  class="h-14 bg-transparent outline-0 cursor-pointer w-full h-14 bg-teal-400 text-slate-100 border-0 rounded-xl text-base text-semibold tracking-wide"
>
  <Icon name={icon} />
  <span class="ml-4">Continue with {brandName}</span>
</button>