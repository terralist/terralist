<script lang="ts">
  import { validateEntry, type FormEntry } from '../../lib/form';

  import Input from '../inputs/Input.svelte';
  import KeyboardAction from '../inputs/KeyboardAction.svelte';

  import Modal from './Modal.svelte'

  export let id: string = Math.random().toString();
  export let title: string;
  export let enabled: boolean = false;
  export let onClose: () => void = () => { };
  export let onSubmit: (entries: Map<string, any>) => void = () => { };
  export let entries: FormEntry[] = [];

  let ref: HTMLFormElement;
  let entriesRefs: any[] = Array.from({ length: entries.length }, () => undefined);
  let entriesErrors: string[] = Array.from({ length: entries.length }, () => "");

  const reset = () => {
    // Disable all highlights and errors
    entriesRefs.forEach(ref => ref.highlight("none"));
    entriesErrors.forEach((_, i) => {entriesErrors[i] = "";});

    // Reset the form
    ref.reset();

    // Put back values for disabled inputs
    entriesRefs
      .filter((_, i) => entries[i].disabled)
      .forEach((ref, i) => ref.setValue(entries[i].value));
  };

  const close = () => {
    reset();
    enabled = false;
    onClose();
  };

  const submit = () => {
    entries.forEach((entry: FormEntry, index: number) => {
      entry.value = entriesRefs[index].value;

      let result = validateEntry(entry);
      
      if (!result.passed) {
        entriesRefs[index].highlight("error");
        entriesErrors[index] = result.message;
      } else {
        if (entry.value.length > 0) {
          entriesRefs[index].highlight("success");
        }

        entriesErrors[index] = "";
      }

      return result.passed;
    });

    if (!entriesErrors.every(e => !e)) {
      return false;
    }

    onSubmit(
      new Map(entries.map((entry: FormEntry, index: number) => [entry.id, entriesRefs[index].value]))
    );

    close();
  }
</script>

{#if enabled}
  <KeyboardAction trigger={'Enter'} action={submit} />
{/if}

<Modal title={title} enabled={enabled} onClose={close}>
  <span slot="body">
    <form id={id} on:submit={submit} class="p-2 w-full grid grid-cols-4 gap-4" bind:this={ref}>
      {#each entries as entry, index}
        <label for={entry.id} class="pt-2 {entry.disabled ? 'italic' : ''}">
          {entry.name}
          {#if entry.required}
            <span class="text-red-700 dark:text-red-200 text-sm text-light">
              *
            </span>
          {/if}
        </label>
        <div class="col-span-3">
          <Input 
            id={entry.id}
            type={entry.type}
            value={entry.value}
            disabled={entry.disabled}
            bind:this={entriesRefs[index]}
          />
          {#if entriesErrors[index]}
          <span class="text-red-700 dark:text-red-200 text-medium">
            {entriesErrors[index]}
          </span>
          {/if}
        </div>
      {/each}
    </form>
  </span>
  <span slot="footer" class="w-full flex justify-end items-center gap-4">
    <button on:click={reset} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-cyan-400 shadow rounded-lg hover:bg-cyan-500 focus:ring-4 focus:outline-none focus:ring-blue-300 dark:bg-cyan-700 dark:hover:bg-cyan-800 dark:focus:ring-blue-700">
      <span class="text-sm text-light uppercase">
        Reset
      </span>
    </button>
    <button on:click={submit} class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      <span class="text-sm text-light uppercase">
        Continue
      </span>
    </button>
  </span>
</Modal>


