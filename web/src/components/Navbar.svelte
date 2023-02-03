<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { Unsubscriber } from "svelte/store";
  import { link } from "svelte-spa-router";
  import { clickOutside } from 'svelte-use-click-outside';

  import Icon from "./Icon.svelte";

  import Searchbar from "./Searchbar.svelte";
  import NavbarAnchor from './NavbarAnchor.svelte';
  import NavbarButton from './NavbarButton.svelte';

  import config from "@/config";
  import context, { type Theme } from "@/context";

  import { useFlag } from '@/lib/hooks';

  const [open, setOpen, resetOpen] = useFlag(false);

  const toggle = () => {
    ($open ? resetOpen : setOpen)();
  };

  let currentTheme: Theme = undefined;
  let themeUnsubscriber: Unsubscriber;

  onMount(() => {
    themeUnsubscriber = context.theme.subscribe((value) => {
      currentTheme = value;
    })
  })

  onDestroy(() => {
    themeUnsubscriber();
  });

  const toggleTheme = () => {
    context.setTheme(currentTheme === "light" ? "dark" : "light");
  };

</script>

<header
  class="fixed z-1 top-0 left-0 flex flex-col lg:flex-row items-center justify-center lg:justify-start lg:pl-4 w-full h-32 lg:h-16 bg-teal-400 dark:bg-teal-700 text-slate-600 dark:text-slate-200 box-border shadow"
>
  <button 
    class="absolute top-0 left-0 grid place-items-center w-16 h-16 lg:hidden" 
    on:click={toggle}
  >
    <Icon name="menu" />
  </button>

  <h1 class="m-0 text-base lg:justify-self-start lg:mr-auto">
    <a href="/" use:link>Terralist v{config.build.env.TERRALIST_VERSION}</a>
  </h1>

  <Searchbar />

  <nav
    class="
      fixed
      z-3
      top-0
      left-0
      w-48
      h-full
      p-5
      text-teal-50
      lg:text-inherit
      lg:justify-self-end
      lg:ml-auto
      flex
      gap-2
      flex-col
      items-start
      bg-zinc-900
      transition
      translate
      duration-300
      lg:transition-none
      lg:static
      lg:translate-x-0
      lg:w-auto
      lg:bg-transparent
      lg:flex-row
      lg:visible
      mb-2
      {$open ? 'translate-x-0 visible' : 'invisible -translate-x-full'}
    "
    use:clickOutside={resetOpen}
  >
    <NavbarAnchor title="Dashboard" href="/" icon="home" />
    <NavbarAnchor title="Settings" href="/settings" icon="settings" />
    <NavbarAnchor title="Sign Out" href="/logout" icon="logout" />
    {#if currentTheme === "dark"}
      <NavbarButton
        title="Light Mode"
        tooltip="Change theme"
        iconClass="solid"
        icon="sun"
      
        onClick={toggleTheme}
      />
    {:else}
      <NavbarButton
        title="Dark Mode"
        tooltip="Change theme"
        iconClass="solid"
        icon="moon"
      
        onClick={toggleTheme}
      />
    {/if}
  </nav>
</header>
