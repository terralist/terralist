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
  baseURL: "/v1/api/artifact",
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

  getOne: (slug: string) => client
    .get<Artifact>(`/${slug}`)
    .then(handleResponse<Artifact>)
    .catch(handleError),

  getAllVersionsForOne: (slug: string) => client
    .get<ArtifactVersion[]>(`/${slug}/version`)
    .then(handleResponse<ArtifactVersion[]>)
    .then(sortVersions)
    .catch(handleError),

  delete: (slug: string, version: string) => client
    .delete<boolean>(`/${slug}/version/${version}`)
    .then(handleResponse<boolean>)
    .catch(handleError)
};

const Artifacts = {
  getAll: async () => await actions.getAll(),
  getOne: async (slug: string) => await actions.getOne(slug),
  getAllVersionsForOne: async (slug: string) => await actions.getAllVersionsForOne(slug),
  delete: async (slug: string, version: string) => await actions.delete(slug, version),
};

export {
  type Artifact,
  type ArtifactVersion,
  Artifacts
};