<script lang="ts">
  import Icon from './Icon.svelte';

  import FormModal from './FormModal.svelte';
  import ErrorModal from './ErrorModal.svelte';

  import Authority from './Authority.svelte';

  import { Authorities, type Authority as AuthorityT } from '@/api/authorities';

  import {
    StringMinimumLengthValidation,
    URLValidation
  } from '@/lib/validation';
  import { useFlag, useQuery } from '@/lib/hooks';
  import { writable } from 'svelte/store';

  const result = useQuery(Authorities.getAll);

  $: authorities = writable($result.data ?? []);
  $: errorMessage = writable($result.error);

  const [createModalEnabled, showCreateModal, hideCreateModal] = useFlag(false);

  const onAuthorityCreateSubmit = async (
    entries: Map<string, string | string[] | undefined>
  ) => {
    const nameValue = entries.get('name');
    const name = Array.isArray(nameValue) ? nameValue.at(0) : nameValue;

    const policyUrlValue = entries.get('policyUrl');
    const policyUrl = Array.isArray(policyUrlValue)
      ? policyUrlValue.at(0)
      : policyUrlValue;

    let result = await Authorities.create(name ?? '', policyUrl ?? '');

    if (result.status === 'OK') {
      authorities.set([...($authorities ?? []), result.data]);
    } else {
      errorMessage.set(result.message);
    }
  };

  const onAuthorityUpdateSubmit = async (_: string, authority: AuthorityT) => {
    let result = await Authorities.update(authority);

    if (result.status === 'OK') {
      authorities.set([
        ...($authorities?.map(a => (a.id == authority.id ? authority : a)) ??
          [])
      ]);
    } else {
      errorMessage.set(result.message);
    }
  };

  const onAuthorityDeleteSubmit = async (id: string) => {
    if (!$authorities) {
      errorMessage.set(`No authorities fetched. Don't know what to delete.`);
      return;
    }

    const authority = $authorities.find((a: AuthorityT) => a.id === id);
    if (!authority) {
      errorMessage.set(`Could not select authority with ID: ${id}`);
      return;
    }

    let result = await Authorities.delete(authority.id);

    if (result.status === 'OK') {
      authorities.set([...$authorities.filter(a => a.id !== id)]);
    } else {
      errorMessage.set(result.message);
    }
  };
</script>

<main class="mt-36 mx-4 lg:mt-14 lg:mx-10 text-slate-600 dark:text-slate-200">
  <section class="mt-20 lg:mx-20">
    {#if $result.isLoading}
      <p>Loading...</p>
    {:else if ($authorities ?? []).length > 0}
      <div
        class="w-full p-2 px-6 grid grid-cols-6 lg:grid-cols-10 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
        <span class="col-span-2 lg:col-span-6"> Name </span>
        <span> Policy </span>
        <span> Signing Keys </span>
        <span> API Keys </span>
        <span class="place-self-end"> Actions </span>
      </div>
      {#each $authorities as authority (authority.id)}
        <Authority
          {authority}
          onUpdate={onAuthorityUpdateSubmit}
          onDelete={onAuthorityDeleteSubmit} />
      {/each}
    {/if}
    <div
      class="w-full px-6 grid grid-cols-6 lg:grid-cols-10 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span class="col-span-4 flex">
        <button
          on:click={showCreateModal}
          class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
          <Icon name="plus" />
          <span class="ml-2 text-sm text-light uppercase"> New Authority </span>
        </button>
      </span>
    </div>
  </section>

  {#if $errorMessage}
    <ErrorModal message={$errorMessage} />
  {/if}

  <FormModal
    title="New authority"
    enabled={$createModalEnabled}
    onClose={hideCreateModal}
    onSubmit={onAuthorityCreateSubmit}
    entries={[
      {
        id: 'name',
        name: 'Name',
        required: true,
        type: 'text',
        validations: [StringMinimumLengthValidation(4)]
      },
      {
        id: 'policyUrl',
        name: 'Policy',
        type: 'text',
        validations: [URLValidation()]
      }
    ]} />
</main>
