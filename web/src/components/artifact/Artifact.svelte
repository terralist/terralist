<script lang="ts">
  import { onMount } from 'svelte';

  import config from "../../config";
  import { indent } from "../../lib/utils";

  import Icon from "../icons/Icon.svelte";
  import Dropdown from "../inputs/Dropdown.svelte";

  import { fetchArtifact, fetchModuleVersions, fetchProviderVersions } from '../../api/artifacts';

  export let type: "module" | "provider";
  export let namespace: string;
  export let name: string;
  export let provider: string = "";
  export let version: string = "";

  const moduleTemplate = `
    module "${name}" {
      source  = "${config.runtime.env.TERRALIST_CANONICAL_DOMAIN}/${namespace}/${name}/${provider}"
      version = "${version}"
    }
  `;

  const providerTemplate = `
    terraform {
      required_providers {
        ${name} = {
          source = "${config.runtime.env.TERRALIST_CANONICAL_DOMAIN}/${namespace}/${name}"
          version = "${version}"
        }
      }
    }

    provider "${name}" {
      # Configuration options
    }
  `;

  const template = indent({
    s: type === "module" ? moduleTemplate : providerTemplate,
    n: 4,
    reverse: true,
  });
  
  const onOptionSelect = (option: string) => {
    const url = [namespace, name].concat(type === 'module' ? [provider] : []).concat(option).join("/").toLowerCase();
    // redirect(url);
    console.log(`Redirect to ${url}`);
  };

  let versions: string[] = [];
  let label: string = version;

  onMount(() => {
    if (type === "module") {
      versions = fetchModuleVersions(namespace, name, provider);
    } else {
      versions = fetchProviderVersions(namespace, name);
    }

    if (version === versions[0]) {
      label = `${version} (latest)`
    }
  });
</script>


<main class="mt-36 mx-4 lg:mt-14 lg:mx-10 text-slate-600 dark:text-slate-200">
  <section class="mt-20 lg:mx-20 flex flex-col gap-8">
    <div class="flex flex-col lg:flex-row justify-between items-start gap-8 lg:items-center">
      <div class="flex justify-start items-center gap-10 mb-4">
        <div class="flex flex-col justify-center items-center dark:text-white">
          <Icon 
            name={type === 'provider' ? 'cloud' : 'hammer'}
            width="8rem"
            height="8rem"
          />
          <span class="text-xs text-zinc-800 dark:text-white">
            {`(${type})`}
          </span>
        </div>
        <div class="flex flex-col justify-center items-start">
          <h2 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white break-words">
            {name + (type === 'module' ? ` (${provider})` : '')}
          </h2>
          <h3 class="text-zinc-800 dark:text-zinc-100">
            @{namespace}
          </h3>
        </div>
      </div>
      <div class="w-full lg:w-auto">
        <Dropdown label={label} options={versions} onSelect={onOptionSelect} />
      </div>
    </div>
    <div class="bg-gray-100 dark:bg-gray-800 border border-teal-400 dark:border-teal-600 p-4 flex flex-col gap-4">
      <h2 class="text-lg font-bold">Usage</h2>
      <p class="text-xs">
        Copy and paste into your Terraform configuration, insert the variables, and run <code>terraform init</code>:
      </p>
      <pre class="bg-gray-200 dark:bg-gray-700 border border-slate-400 dark:border-slate-600 p-3 text-xs overflow-y-auto">{template}</pre>
    </div>
  </section>
</main>