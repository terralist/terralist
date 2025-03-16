import type { Artifact } from '@/api/artifacts';

type LocatableArtifactBase = {
  type: 'module' | 'provider';
  namespace: string;
  name: string;
  provider?: string | never;
  version: string;
};

type LocatableProvider = LocatableArtifactBase & {
  type: 'provider';
  provider?: never;
};

type LocatableModule = LocatableArtifactBase & {
  type: 'module';
  provider: string;
};

type LocatableArtifact = LocatableProvider | LocatableModule;

function isArtifact(arg: unknown): arg is Artifact {
  return (
    (arg as Artifact)?.id != undefined &&
    typeof (arg as Artifact).id == 'string'
  );
}

const computeArtifactUrl = (artifact: LocatableArtifact | Artifact): string => {
  if (isArtifact(artifact)) {
    const { type, namespace, name, provider, versions } = artifact;

    artifact = {
      type,
      namespace,
      name,
      provider,
      version: versions[0]
    } as LocatableArtifact;
  }

  if (artifact.type == 'module' && !artifact.provider) {
    throw new Error("Invalid module definition: missing 'provider' field.");
  }

  const slug = [artifact.namespace, artifact.name]
    .concat(artifact.type == 'module' ? [artifact.provider] : [])
    .concat(artifact.version)
    .join('/')
    .toLowerCase();

  const category = {
    module: 'modules',
    provider: 'providers'
  }[artifact.type];

  return `/${category}/${slug}`;
};

export { type LocatableArtifact, computeArtifactUrl };
