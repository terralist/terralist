<script lang="ts">
  import { onMount } from "svelte";

  export let trigger: string;
  export let action: (keys: Set<string>) => void;
  export let preventDefault: boolean = false;

  let keys: Set<string> = new Set();
  let keysPressed: Set<string> = new Set();

  onMount(() => {
    keys = new Set(trigger.split("+").map(key => key.trim()));
  });

  const onKeyUp = (e: KeyboardEvent) => {
    if (keys.has(e.key)) {
      if (preventDefault) {
        e.preventDefault();
      }

      keysPressed.add(e.key);
    
      if (keys.size == keysPressed.size && [...keysPressed].every(k => keys.has(k))) {
        action(keysPressed);
      }
    }
  };

  const onKeyDown = (e: KeyboardEvent) => {
    if (keysPressed.has(e.key)) {
      keysPressed.delete(e.key);
    }
  };
</script>

<svelte:window 
  on:keydown={onKeyDown}
  on:keyup={onKeyUp}
/>

