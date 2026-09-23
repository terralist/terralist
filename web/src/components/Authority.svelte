<script lang="ts">
  import TransparentButton from './TransparentButton.svelte';
  import CaretButton from './CaretButton.svelte';
  import Icon from './Icon.svelte';

  import FormModal from './FormModal.svelte';
  import ConfirmationModal from './ConfirmationModal.svelte';
  import ErrorModal from './ErrorModal.svelte';

  import Key from './Key.svelte';

  import type { Authority as AuthorityT } from '@/api/authorities';
  import { Keys, type Key as KeyT } from '@/api/keys';

  import { URLValidation } from '@/lib/validation';
  import { useFlag, useToggle } from '@/lib/hooks';

  export let authority: AuthorityT;
  export let onUpdate: (id: string, authority: AuthorityT) => void = () => {};
  export let onDelete: (id: string) => void = () => {};

  let errorMessage: string = '';

  const [createKeyModalEnabled, showCreateKeyModal, hideCreateKeyModal] =
    useFlag(false);
  const [updateModalEnabled, showUpdateModal, hideUpdateModal] = useFlag(false);
  const [deleteModalEnabled, showDeleteModal, hideDeleteModal] = useFlag(false);

  const [showKeys, toggleShowKeys] = useToggle(false);

  const update = (entries: Map<string, string | string[] | undefined>) => {
    const policyUrlValue = entries.get('policyUrl');
    const policyUrl = Array.isArray(policyUrlValue)
      ? policyUrlValue.at(0)
      : policyUrlValue;

    const publicValue = entries.get('public');
    const isPublic = Array.isArray(publicValue)
      ? publicValue.at(0)
      : publicValue;

    onUpdate(authority.id, {
      ...authority,
      policyUrl: policyUrl ?? '',
      public: (isPublic ?? 'false') == 'true'
    });
  };

  const remove = () => {
    onDelete(authority.id);
  };

  const createKeySubmit = async (
    entries: Map<string, string | string[] | undefined>
  ) => {
    const keyIdValue = entries.get('key_id');
    const keyId = Array.isArray(keyIdValue) ? keyIdValue.at(0) : keyIdValue;

    const asciiArmorValue = entries.get('ascii_armor');
    const asciiArmor = Array.isArray(asciiArmorValue)
      ? asciiArmorValue.at(0)
      : asciiArmorValue;

    const trustSignatureValue = entries.get('trust_signature');
    const trustSignature = Array.isArray(trustSignatureValue)
      ? trustSignatureValue.at(0)
      : trustSignatureValue;

    let result = await Keys.create(
      authority.id,
      keyId ?? '',
      asciiArmor ?? '',
      trustSignature ?? ''
    );

    if (result.status === 'OK') {
      authority.keys = [...authority.keys, result.data];
    } else {
      errorMessage = result.message;
    }
  };

  const onKeyDelete = async (id: string) => {
    const key = authority.keys.find((k: KeyT) => k.id === id);
    if (!key) {
      errorMessage = `Could not select key with ID: ${id}.`;
      return;
    }

    let result = await Keys.delete(authority.id, key.id);

    if (result.status === 'OK') {
      authority.keys = [...authority.keys.filter((k: KeyT) => k.id !== id)];
    } else {
      errorMessage = result.message;
    }
  };
</script>

<div class="mb-4">
  <div
    class="w-full rounded-lg p-2 px-6 bg-teal-400 dark:bg-teal-700 grid grid-cols-6 lg:grid-cols-10 place-items-start">
    <span class="col-span-2 lg:col-span-6">{authority.name}</span>
    <span>
      {#if authority.policyUrl}
        <a href={authority.policyUrl} target="_blank" rel="noreferrer">
          <TransparentButton>
            <Icon name="external-link" />
          </TransparentButton>
        </a>
      {:else}
        <span
          class="p-1.5 flex flex-col md:flex-row justify-center items-center">
          <Icon name="close-square" />
        </span>
      {/if}
    </span>
    <span class="p-1.5 flex flex-col md:flex-row justify-center items-center">
      {#if authority.public}
        <Icon name="checkmark-square" />
      {:else}
        <Icon name="close-square" />
      {/if}
    </span>
    <span class="flex flex-col md:flex-row justify-center items-center">
      <TransparentButton onClick={showCreateKeyModal}>
        <Icon name="plus" />
      </TransparentButton>
      <span class="ml-0 md:ml-2">
        {authority.keys?.length ?? 0}
      </span>
      {#if authority.keys?.length > 0}
        <CaretButton
          class="ml-0 md:ml-2"
          onClick={toggleShowKeys}
          enabled={$showKeys} />
      {/if}
    </span>
    <span class="place-self-end flex justify-center items-center">
      <TransparentButton onClick={showUpdateModal}>
        <Icon name="edit-box" />
      </TransparentButton>
      <TransparentButton onClick={showDeleteModal}>
        <Icon name="trash" />
      </TransparentButton>
    </span>
  </div>
  {#if $showKeys}
    <div
      class="w-full p-2 px-6 grid grid-cols-4 lg:grid-cols-8 place-items-start text-xs lg:text-sm text-light uppercase text-zinc-500 dark:text-zinc-200">
      <span class="lg:col-span-5"> Key ID </span>
      <span> ASCII Armor </span>
      <span> Trust Signature </span>
      <span class="place-self-end"> Actions </span>
    </div>
    {#each authority.keys as key (key.id)}
      <Key
        authorityKey={key}
        authorityName={authority.name}
        isAlone={authority.keys.length === 1}
        onDelete={onKeyDelete} />
    {/each}
  {/if}

  <FormModal
    title={`Update authority ${authority.name}`}
    enabled={$updateModalEnabled}
    onClose={hideUpdateModal}
    onSubmit={update}
    entries={[
      {
        id: 'name',
        name: 'Name',
        type: 'text',
        disabled: true,
        value: authority.name
      },
      {
        id: 'policyUrl',
        name: 'Policy',
        type: 'text',
        value: authority.policyUrl,
        validations: [URLValidation()]
      },
      {
        id: 'public',
        name: 'Public',
        type: 'checkbox',
        value: authority.public ? 'true' : 'false'
      }
    ]} />

  <ConfirmationModal
    title={`Remove authority ${authority.name}`}
    enabled={$deleteModalEnabled}
    onClose={hideDeleteModal}
    onSubmit={remove}>
    Removing the <b>{authority.name}</b> authority will also remove all
    artifacts uploaded to the <b>{authority.name}</b> namespace.
    <br /><br />
    Are you sure?
  </ConfirmationModal>

  <FormModal
    title={`Add a new authority key to ${authority.name}`}
    enabled={$createKeyModalEnabled}
    onClose={hideCreateKeyModal}
    onSubmit={createKeySubmit}
    entries={[
      {
        id: 'key_id',
        name: 'Key ID',
        required: true,
        type: 'text',
        validations: []
      },
      {
        id: 'ascii_armor',
        name: 'ASCII Armor',
        type: 'textarea',
        validations: []
      },
      {
        id: 'trust_signature',
        name: 'Trust Signature',
        type: 'textarea',
        validations: []
      }
    ]} />

  {#if errorMessage}
    <ErrorModal bind:message={errorMessage} />
  {/if}
</div>
