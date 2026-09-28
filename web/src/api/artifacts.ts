import {
  createClient,
  handleResponse,
  handleError,
  type Result
} from '@/api/api.utils';

import cmp from 'semver-compare';

type ArtifactVersion = string;

type VersionOrigin = 'manual' | 'upstream';

type VersionDetails = {
  version: ArtifactVersion;
  origin: VersionOrigin;
  // mirrorOnly marks a provider version served by the network mirror only,
  // for lack of a signed SHA256SUMS file.
  mirrorOnly?: boolean;
};

// FetchResults reports, per platform of a provider, the outcome of fetching a
// version from the upstream; a module fetch reports nothing.
type FetchResults = {
  results?: { platform: string; error?: string }[];
};

type ArtifactVersions = {
  versions: VersionDetails[];
  canDelete: boolean;
  canFetch: boolean;
  // canBlock allows denying a pulled version by a rule, and deleting it.
  canBlock: boolean;
};

type Submodule = {
  path: string;
};

type ArtifactVersionWithDocumentation = {
  version: ArtifactVersion;
  documentation?: string;
  submodules?: Submodule[];
};

type ArtifactBase = {
  id: string;
  fullName: string;
  namespace: string;
  name: string;
  versions: ArtifactVersion[];
  createdAt: Date;
  updatedAt: Date;
};

type ProviderArtifact = ArtifactBase & {
  type: 'provider';
  provider?: never;
};

type ModuleArtifact = ArtifactBase & {
  type: 'module';
  provider: string;
};

type Artifact = ProviderArtifact | ModuleArtifact;

const createDateAttributes = (artifact: Artifact): Artifact => {
  return {
    ...artifact,
    createdAt: new Date(artifact.createdAt),
    updatedAt: new Date(artifact.updatedAt)
  };
};

const client = createClient({
  baseURL: '/v1/api/artifacts',
  timeout: 120000
});

const setDateAttributes = <T extends Artifact | Artifact[]>(
  r: Result<T>
): Result<T> => {
  const { data, ...rest } = r;

  if (!data) {
    return r;
  }

  if (Array.isArray(data)) {
    return {
      data: data.map(createDateAttributes),
      ...rest
    } as Result<T>;
  }

  return {
    data: createDateAttributes(data),
    ...rest
  } as Result<T>;
};

const sortArtifactsVersions = (r: Result<Artifact[]>): Result<Artifact[]> => {
  const { data: artifacts, ...rest } = r;

  const result = artifacts?.map(a => {
    return {
      ...a,
      versions: a.versions.sort(cmp).reverse()
    };
  });

  return {
    data: result,
    ...rest
  } as Result<Artifact[]>;
};

const sortVersions = (
  r: Result<ArtifactVersions>
): Result<ArtifactVersions> => {
  if (r.status == 'ERROR') {
    return r;
  }

  const { data, ...rest } = r;

  return {
    data: {
      ...data,
      versions: data.versions.sort((a, b) => cmp(b.version, a.version))
    },
    ...rest
  };
};

const actions = {
  getAll: async () =>
    client
      .get<Artifact[]>('/')
      .then(handleResponse<Artifact[]>)
      .then(setDateAttributes)
      .then(sortArtifactsVersions)
      .catch(handleError),

  getOne: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ) =>
    client
      .get<Artifact>([namespace, name, provider].filter(e => e).join('/'))
      .then(handleResponse<Artifact>)
      .then(setDateAttributes)
      .catch(handleError),

  getAllVersionsForOne: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ) =>
    client
      .get<ArtifactVersions>(
        `/${[namespace, name, provider].filter(e => e).join('/')}/version`
      )
      .then(handleResponse<ArtifactVersions>)
      .then(sortVersions)
      .catch(handleError),

  getOneVersion: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string
  ) =>
    client
      .get<ArtifactVersionWithDocumentation>(
        `/${[namespace, name, provider].filter(e => e).join('/')}/version/${version}`
      )
      .then(handleResponse<ArtifactVersionWithDocumentation>)
      .catch(handleError),

  fetchFromUpstream: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string,
    platforms: string[]
  ) =>
    createClient({ baseURL: '/v1/api', timeout: 600000 })
      .post<FetchResults>(
        provider
          ? `/modules/${namespace}/${name}/${provider}/${version}/fetch`
          : `/providers/${namespace}/${name}/${version}/fetch`,
        provider ? {} : { platforms }
      )
      .then(handleResponse<FetchResults>)
      .catch(handleError),

  delete: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string
  ) =>
    client
      .delete<boolean>(
        `/${[namespace, name, provider].filter(e => e).join('/')}/version/${version}`
      )
      .then(handleResponse<boolean>)
      .catch(handleError),

  getSubmoduleDocumentation: async (
    namespace: string,
    name: string,
    provider: string,
    version: string,
    submodulePath: string
  ) =>
    client
      .get<{ documentation: string }>(
        `/modules/${namespace}/${name}/${provider}/${version}/submodules/${submodulePath}`,
        {
          baseURL: '/v1/api'
        }
      )
      .then(handleResponse<{ documentation: string }>)
      .catch(handleError)
};

const Artifacts = {
  getAll: async () => await actions.getAll(),
  getOne: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ) => await actions.getOne(namespace, name, provider),
  getSubmoduleDocumentation: async (
    namespace: string,
    name: string,
    provider: string,
    version: string,
    submodulePath: string
  ) =>
    await actions.getSubmoduleDocumentation(
      namespace,
      name,
      provider,
      version,
      submodulePath
    ),
  getAllVersionsForOne: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ) => await actions.getAllVersionsForOne(namespace, name, provider),
  getOneVersion: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string
  ) => await actions.getOneVersion(namespace, name, provider, version),
  delete: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string
  ) => await actions.delete(namespace, name, provider, version),
  fetchFromUpstream: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string,
    platforms: string[] = []
  ) =>
    await actions.fetchFromUpstream(
      namespace,
      name,
      provider,
      version,
      platforms
    )
};

export {
  type Artifact,
  type ArtifactVersion,
  type ArtifactVersions,
  type VersionDetails,
  type VersionOrigin,
  type ArtifactVersionWithDocumentation,
  type Submodule,
  Artifacts
};
