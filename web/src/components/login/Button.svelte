<script lang="ts">
  export let brandName: String;
  export let iconClass: String = "brands";
  export let icon: String = brandName.toLowerCase();
</script>

<button
  class="h-14 bg-transparent outline-0 cursor-pointer w-full h-14 bg-teal-400 text-slate-100 border-0 rounded-xl text-base text-semibold tracking-wide"
>
  <i class="fa-{iconClass} fa-{icon}"></i>
  <span class="ml-4">Continue with {brandName}</span>
</button>