import type { Artifact } from "@/api/artifacts";

const computeArtifactUrl = (artifact: Artifact) => {
  const slug = [artifact.namespace, artifact.name]
      .concat(artifact.type === 'module' ? [artifact.provider] : [])
      .concat(artifact.versions[0])
      .join("/")
      .toLowerCase();
  
  const category = {
    "module": "modules",
    "provider": "providers",
  }[artifact.type];

  return `/${category}/${slug}`;
};

export {
  computeArtifactUrl
};