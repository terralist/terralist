<script lang="ts">
  import { onMount } from "svelte";
  import { clickOutside } from 'svelte-use-click-outside';

  import { defaultIfNull } from '../../lib/utils';

  import Searchbar from "./Searchbar.svelte";
  import Anchor from './Anchor.svelte';

  let open = false;

  let darkMode: boolean = defaultIfNull(JSON.parse(localStorage.getItem("preferred.theme.darkMode")), false);

  const toggle = () => {
    open = !open
  };

  const toggleTheme = () => {
    darkMode = !darkMode;
    localStorage.setItem("preferred.theme.darkMode", JSON.stringify(darkMode));
    setTheme();
  };

  const setTheme = () => {
    const root = document.documentElement;

    if (darkMode) {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  };

  onMount(() => {
		setTheme();
	});
</script>

<header
  class="fixed z-1 top-0 left-0 flex flex-col lg:flex-row items-center justify-center lg:justify-start lg:pl-4 w-full h-32 lg:h-16 bg-teal-400 dark:bg-teal-700 text-slate-600 dark:text-slate-200 box-border shadow"
>
  <button 
    class="absolute top-0 left-0 grid place-items-center w-16 h-16 lg:hidden" 
    on:click={toggle}
  >
    <i class="fa fa-bars"></i>
  </button>

  <h1 class="m-0 text-base lg:justify-self-start lg:mr-auto">
    <a href="index.html">Terralist v0.1.0</a>
  </h1>

  <Searchbar />

  <nav
    class="fixed z-3 top-0 left-0 w-48 h-full p-5 text-teal-50 lg:text-inherit lg:justify-self-end lg:ml-auto flex gap-2 flex-col items-start bg-zinc-900 transition translate duration-300 lg:transition-none lg:static lg:translate-x-0 lg:w-auto lg:bg-transparent lg:flex-row lg:visible {open ? 'translate-x-0 visible' : 'invisible -translate-x-full'}"
    use:clickOutside={() => {open = false}}
  >
    <Anchor title="Manage" icon="gear" className="pt-0.5 lg:pt-0" />
    <Anchor title="Sign Out" icon="arrow-right-from-bracket" />
    {#if darkMode}
      <Anchor 
        title="Light Mode"
        tooltip="Change theme"
        iconClass="solid"
        icon="sun"
        className="pt-0.5 lg:pt-0"
        clickHandler={toggleTheme}
      />
    {:else}
      <Anchor 
        title="Dark Mode"
        tooltip="Change theme"
        iconClass="solid"
        icon="moon"
        className="pt-0.5 lg:pt-0"
        clickHandler={toggleTheme}
      />
    {/if}
  </nav>
</header>
