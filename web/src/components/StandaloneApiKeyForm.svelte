<script lang="ts">
  import Icon from './Icon.svelte';
  import TransparentButton from './TransparentButton.svelte';
  import Modal from './Modal.svelte';

  import type {
    CreatePolicyDTO,
    CreateStandaloneApiKeyDTO
  } from '@/api/standaloneApiKeys';

  export let enabled: boolean = false;
  export let onClose: () => void = () => {};
  export let onSubmit: (dto: CreateStandaloneApiKeyDTO) => void = () => {};
  export let authorities: string[] = [];

  const resources = ['modules', 'providers', 'authorities', 'api-keys', '*'];
  const actions = ['get', 'create', 'update', 'delete', '*'];
  const effects = ['allow', 'deny'];

  type PolicyRow = {
    resource: string;
    action: string;
    effect: string;
    // Structured object fields
    authority: string;
    module: string;
    provider: string;
    apiKey: string;
  };

  let name = '';
  let expireIn = 0;
  let policies: PolicyRow[] = [emptyPolicy()];
  let error = '';

  function emptyPolicy(): PolicyRow {
    return {
      resource: 'modules',
      action: 'get',
      effect: 'allow',
      authority: '*',
      module: '*',
      provider: '*',
      apiKey: '*'
    };
  }

  function addPolicy() {
    policies = [...policies, emptyPolicy()];
  }

  function removePolicy(index: number) {
    policies = policies.filter((_, i) => i !== index);
  }

  function buildObject(policy: PolicyRow): string {
    switch (policy.resource) {
      case 'modules':
        return `${policy.authority}/${policy.module}/${policy.provider}`;
      case 'providers':
        return `${policy.authority}/${policy.provider}`;
      case 'authorities':
        return policy.authority;
      case 'api-keys':
        return policy.apiKey;
      case '*':
        return '*';
      default:
        return '*';
    }
  }

  function reset() {
    name = '';
    expireIn = 0;
    policies = [emptyPolicy()];
    error = '';
  }

  function close() {
    reset();
    enabled = false;
    onClose();
  }

  function submit() {
    error = '';

    if (!name || name.length < 4) {
      error = 'Name must be at least 4 characters.';
      return;
    }

    if (policies.length === 0) {
      error = 'At least one policy is required.';
      return;
    }

    const dto: CreateStandaloneApiKeyDTO = {
      name,
      expireIn,
      policies: policies.map(
        (p): CreatePolicyDTO => ({
          resource: p.resource,
          action: p.action,
          object: buildObject(p),
          effect: p.effect
        })
      )
    };

    onSubmit(dto);
    close();
  }

  const inputClass =
    'w-full px-2 h-8 bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200 shadow border-none text-sm rounded-lg focus:ring-0 outline-none';

  const selectClass =
    'w-full px-1 h-8 bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200 shadow border-none text-sm rounded-lg focus:ring-0 outline-none cursor-pointer';
</script>

<Modal title="New API Key" {enabled} onClose={close}>
  <span slot="body">
    <div class="space-y-4">
      {#if error}
        <div class="text-sm text-red-600 dark:text-red-400">{error}</div>
      {/if}

      <div class="grid grid-cols-4 gap-4 items-center">
        <label for="ak-name" class="text-sm">
          Name
          <span class="text-red-700 dark:text-red-200 text-sm">*</span>
        </label>
        <div class="col-span-3">
          <input
            id="ak-name"
            type="text"
            class={inputClass}
            placeholder="e.g. ci-deploy-key"
            bind:value={name} />
        </div>

        <label for="ak-expire" class="text-sm">Expires in (hours)</label>
        <div class="col-span-3">
          <input
            id="ak-expire"
            type="number"
            min="0"
            class={inputClass}
            placeholder="0 = never"
            bind:value={expireIn} />
        </div>
      </div>

      <div>
        <div class="flex justify-between items-center mb-2">
          <p class="text-xs uppercase text-zinc-400">Policies</p>
          <button
            on:click={addPolicy}
            class="inline-flex items-center py-1 px-2 text-xs bg-teal-400 dark:bg-teal-700 rounded-lg hover:bg-teal-500 dark:hover:bg-teal-800">
            <Icon name="plus" />
            <span class="ml-1 uppercase">Add</span>
          </button>
        </div>

        {#each policies as policy, index (index)}
          <div
            class="mb-3 p-3 rounded-lg bg-slate-100 dark:bg-slate-800 space-y-2">
            <div class="flex justify-between items-center">
              <span class="text-xs uppercase text-zinc-400"
                >Policy {index + 1}</span>
              {#if policies.length > 1}
                <TransparentButton onClick={() => removePolicy(index)}>
                  <Icon name="close" width="1rem" height="1rem" />
                </TransparentButton>
              {/if}
            </div>

            <div class="grid grid-cols-3 gap-2">
              <div>
                <label for="policy-res-{index}" class="text-xs text-zinc-400"
                  >Resource</label>
                <select
                  id="policy-res-{index}"
                  class={selectClass}
                  bind:value={policy.resource}>
                  {#each resources as r (r)}
                    <option value={r}>{r}</option>
                  {/each}
                </select>
              </div>

              <div>
                <label for="policy-act-{index}" class="text-xs text-zinc-400"
                  >Action</label>
                <select
                  id="policy-act-{index}"
                  class={selectClass}
                  bind:value={policy.action}>
                  {#each actions as a (a)}
                    <option value={a}>{a}</option>
                  {/each}
                </select>
              </div>

              <div>
                <label for="policy-eff-{index}" class="text-xs text-zinc-400"
                  >Effect</label>
                <select
                  id="policy-eff-{index}"
                  class={selectClass}
                  bind:value={policy.effect}>
                  {#each effects as e (e)}
                    <option value={e}>{e}</option>
                  {/each}
                </select>
              </div>
            </div>

            <div>
              <label class="text-xs text-zinc-400">Object</label>
              {#if policy.resource === '*'}
                <div class="grid grid-cols-1 gap-2">
                  <input
                    type="text"
                    class="{inputClass} opacity-50"
                    value="*"
                    disabled />
                </div>
              {:else if policy.resource === 'modules'}
                <div class="grid grid-cols-3 gap-2">
                  <div>
                    <label
                      for="policy-auth-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >Authority</label>
                    <select
                      id="policy-auth-{index}"
                      class={selectClass}
                      bind:value={policy.authority}>
                      <option value="*">* (all)</option>
                      {#each authorities as auth (auth)}
                        <option value={auth}>{auth}</option>
                      {/each}
                    </select>
                  </div>
                  <div>
                    <label
                      for="policy-mod-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >Module</label>
                    <input
                      id="policy-mod-{index}"
                      type="text"
                      class={inputClass}
                      placeholder="* (all)"
                      bind:value={policy.module} />
                  </div>
                  <div>
                    <label
                      for="policy-prov-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >Provider</label>
                    <input
                      id="policy-prov-{index}"
                      type="text"
                      class={inputClass}
                      placeholder="* (all)"
                      bind:value={policy.provider} />
                  </div>
                </div>
              {:else if policy.resource === 'providers'}
                <div class="grid grid-cols-2 gap-2">
                  <div>
                    <label
                      for="policy-auth-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >Authority</label>
                    <select
                      id="policy-auth-{index}"
                      class={selectClass}
                      bind:value={policy.authority}>
                      <option value="*">* (all)</option>
                      {#each authorities as auth (auth)}
                        <option value={auth}>{auth}</option>
                      {/each}
                    </select>
                  </div>
                  <div>
                    <label
                      for="policy-prov-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >Provider</label>
                    <input
                      id="policy-prov-{index}"
                      type="text"
                      class={inputClass}
                      placeholder="* (all)"
                      bind:value={policy.provider} />
                  </div>
                </div>
              {:else if policy.resource === 'authorities'}
                <div class="grid grid-cols-1 gap-2">
                  <div>
                    <label
                      for="policy-auth-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >Authority</label>
                    <select
                      id="policy-auth-{index}"
                      class={selectClass}
                      bind:value={policy.authority}>
                      <option value="*">* (all)</option>
                      {#each authorities as auth (auth)}
                        <option value={auth}>{auth}</option>
                      {/each}
                    </select>
                  </div>
                </div>
              {:else if policy.resource === 'api-keys'}
                <div class="grid grid-cols-1 gap-2">
                  <div>
                    <label
                      for="policy-ak-{index}"
                      class="text-xs text-zinc-500 dark:text-zinc-400"
                      >API Key</label>
                    <input
                      id="policy-ak-{index}"
                      type="text"
                      class={inputClass}
                      placeholder="* (all)"
                      bind:value={policy.apiKey} />
                  </div>
                </div>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>
  </span>
  <span slot="footer" class="w-full flex justify-end items-center gap-4">
    <button
      on:click={reset}
      class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-cyan-400 shadow rounded-lg hover:bg-cyan-500 focus:ring-4 focus:outline-none focus:ring-blue-300 dark:bg-cyan-700 dark:hover:bg-cyan-800 dark:focus:ring-blue-700">
      <span class="text-sm text-light uppercase">Reset</span>
    </button>
    <button
      on:click={submit}
      class="inline-flex justify-center items-center py-2 px-3 text-sm font-medium shadow text-center bg-teal-400 shadow rounded-lg hover:bg-teal-500 focus:ring-4 focus:outline-none focus:ring-green-300 dark:bg-teal-700 dark:hover:bg-teal-800 dark:focus:ring-green-700">
      <span class="text-sm text-light uppercase">Create</span>
    </button>
  </span>
</Modal>
