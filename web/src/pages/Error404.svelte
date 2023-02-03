<script lang="ts">
  import Icon from "@/components/Icon.svelte";
  import Navbar from "@/components/Navbar.svelte";
</script>

<Navbar />
<main class="mt-36 lg:mt-20 mx-10">
  <div class="absolute top-0 left-0 flex justify-center items-center text-center w-screen h-screen -z-10">
    <div class="flex flex-col justify-center items-center gap-4">
      <span class="flex">
        <Icon name="number-4" width="32px" height="32px" class="fill-zinc-900 dark:fill-zinc-100" />
        <Icon name="number-0" width="32px" height="32px" class="fill-zinc-900 dark:fill-zinc-100" />
        <Icon name="number-4" width="32px" height="32px" class="fill-zinc-900 dark:fill-zinc-100" />
      </span>
      <p class="font-medium text-medium text-zinc-900 dark:text-zinc-100">This is not the web page you're looking for.</p>
    </div>
  </div>
</main>
