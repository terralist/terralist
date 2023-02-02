<script lang="ts">
  import { currentPath } from '../../lib/url';
  import { fromPath, type Artifact as ArtifactT } from '../../lib/artifact';

  import Header from '../../components/navbar/Header.svelte';
  import Artifact from '../../components/artifact/Artifact.svelte';

  let artifact: ArtifactT;
  if (import.meta.env.MODE !== "development") {
    // This cannot be invalid, since it must be validated by the server
    const path = currentPath();

    // This will always return a valid artifact, since it is validated by the server
    artifact = fromPath(path);
  } else {
    // On development mode, the path will point to `artifact.html`, so we need to
    // hardcode an artifact
    const isModule: boolean = true;

    artifact = isModule ? {
      type: "module",
      namespace: "HashiCorp",
      name: "vpc",
      provider: "aws",
      version: "9.0.0",
    } : {
      type: "provider",
      namespace: "HashiCorp",
      name: "aws",
      version: "9.0.0",
    };
  }
</script>

<main>
  <Header />
  <Artifact {...artifact} />
</main>
