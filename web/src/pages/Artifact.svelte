<script lang="ts">
  import Navbar from '@/components/Navbar.svelte';
  import Artifact from '@/components/Artifact.svelte';

  export let params: {
    namespace: string,
    name: string,
    provider?: string,
    version?: string
  };
</script>

<Navbar />
{#key params}
  <Artifact
    type={params.provider ? "module" : "provider"}
    namespace={params.namespace.toLowerCase()}
    name={params.name.toLowerCase()}
    provider={params.provider.toLowerCase()}
    version={params.version}
  />
{/key}

