<script lang="ts">
    import Button from "../../components/login/Button.svelte";
    import Logo from "../../components/Logo.svelte";

  
</script>

<main>
  <header class="hidden md:flex h-full w-full pr-96 items-center justify-center">
    <Logo />
  </header>
  <container 
    class="fixed top-50 left-50 -translate-x-1/2 -translate-y-1/2 w-11/12 pt-16 pb-11 px-8 transition-width trainsition-slowest ease rounded-2xl bg-slate-200 grid place-items-center text-center sm:w-96 md:top-0 md:right-0 md:left-auto md:m-0 md:h-full md:rounded-none md:translate-x-0 md:translate-y-0 xl:w-1/4 "
  >
    <Logo className="mb-5 md:hidden" />
    <div class="mb-8">
      <h2 class="text-3xl font-semibold mt-0 mx-0">
        Terralist
        <span class="opacity-40 mb-1.5 font-medium text-base">v0.1.0</span>
      </h2>
      <h3 class="opacity-40 mb-1.5 font-medium text-base">Company</h3>
    </div>
    <section class="grid gap-8 place-items-center w-full m-0">
      <Button brandName="GitHub" />
      <Button brandName="Google" />
      <Button brandName="BitBucket" />
    </section>
  </container>
</main>
