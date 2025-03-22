import {
  createClient,
  handleResponse,
  handleError,
  type Result
} from '@/api/api.utils';

import cmp from 'semver-compare';

type ArtifactVersion = string;

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
  } as Artifact;
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
  r: Result<ArtifactVersion[]>
): Result<ArtifactVersion[]> => {
  if (r.status == 'ERROR') {
    return r;
  }

  const { data: versions, ...rest } = r;

  return {
    data: versions.sort(cmp).reverse(),
    ...rest
  } as Result<ArtifactVersion[]>;
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
      .get<ArtifactVersion[]>(
        `/${[namespace, name, provider].filter(e => e).join('/')}/version`
      )
      .then(handleResponse<ArtifactVersion[]>)
      .then(sortVersions)
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
      .catch(handleError)
};

const Artifacts = {
  getAll: async () => await actions.getAll(),
  getOne: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ) => await actions.getOne(namespace, name, provider),
  getAllVersionsForOne: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ) => await actions.getAllVersionsForOne(namespace, name, provider),
  delete: async (
    namespace: string,
    name: string,
    provider: string | undefined,
    version: string
  ) => await actions.delete(namespace, name, provider, version)
};

export { type Artifact, type ArtifactVersion, Artifacts };
