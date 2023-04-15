import { createClient, handleResponse, handleError, type Result } from "@/api/api.utils";

import cmp from 'semver-compare';

type ArtifactVersion = string;

interface Artifact {
  id: string,
  fullName: string,
  authority: string,
  name: string,
  provider?: string,
  type: "provider" | "module",
  versions?: ArtifactVersion[],
  latest: string,
  createdAt: Date,
  updatedAt: Date,
};

const client = createClient({
  baseURL: "/v1/api/artifacts",
  timeout: 120,
});

const sortVersions = (r: Result<ArtifactVersion[]>): Result<ArtifactVersion[]> => {
  const {data: versions, ...rest} = r;

  return {
    data: versions.sort(cmp).reverse(),
    ...rest,
  } as Result<ArtifactVersion[]>;
};

const actions = {
  getAll: () => client
    .get<Artifact[]>("/")
    .then(handleResponse<Artifact[]>)
    .catch(handleError),

  getOne: (namespace: string, name: string, provider: string | undefined) => client
    .get<Artifact>([namespace, name, provider].filter(e => e).join("/"))
    .then(handleResponse<Artifact>)
    .catch(handleError),

  getAllVersionsForOne: (namespace: string, name: string, provider: string | undefined) => client
    .get<ArtifactVersion[]>(`/${[namespace, name, provider].filter(e => e).join("/")}/version`)
    .then(handleResponse<ArtifactVersion[]>)
    .then(sortVersions)
    .catch(handleError),

  delete: (namespace: string, name: string, provider: string | undefined, version: string) => client
    .delete<boolean>(`/${[namespace, name, provider].filter(e => e).join("/")}/version/${version}`)
    .then(handleResponse<boolean>)
    .catch(handleError)
};

const Artifacts = {
  getAll: async () => await actions.getAll(),
  getOne: async (namespace: string, name: string, provider: string | undefined) => await actions.getOne(namespace, name, provider),
  getAllVersionsForOne: async (namespace: string, name: string, provider: string | undefined) => await actions.getAllVersionsForOne(namespace, name, provider),
  delete: async (namespace: string, name: string, provider: string | undefined, version: string) => await actions.delete(namespace, name, provider, version),
};

export {
  type Artifact,
  type ArtifactVersion,
  Artifacts
};