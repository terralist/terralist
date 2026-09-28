import {
  createClient,
  handleResponse,
  handleError,
  type Result
} from '@/api/api.utils';

type Platform = {
  os: string;
  arch: string;
};

// OfferedVersion is a version the registry protocol lists for an artifact,
// merged with the versions its authority's upstream registry offers when the
// caller may fetch them. Module versions have no platforms.
type OfferedVersion = {
  version: string;
  platforms: Platform[];
};

type ProviderVersions = {
  versions: OfferedVersion[];
};

type ModuleVersions = {
  modules: { versions: { version: string }[] }[];
};

const client = createClient({
  baseURL: '/v1',
  timeout: 120000
});

const mapResult = <T, U>(r: Result<T>, fn: (data: T) => U): Result<U> =>
  r.status === 'OK' ? { ...r, data: fn(r.data) } : r;

const Registry = {
  offeredVersions: async (
    namespace: string,
    name: string,
    provider: string | undefined
  ): Promise<Result<OfferedVersion[]>> => {
    if (provider) {
      return client
        .get<ModuleVersions>(
          `/modules/${namespace}/${name}/${provider}/versions`
        )
        .then(handleResponse<ModuleVersions>)
        .then(r =>
          mapResult(r, data =>
            (data.modules[0]?.versions ?? []).map(v => ({
              version: v.version,
              platforms: []
            }))
          )
        )
        .catch(handleError);
    }

    return client
      .get<ProviderVersions>(`/providers/${namespace}/${name}/versions`)
      .then(handleResponse<ProviderVersions>)
      .then(r => mapResult(r, data => data.versions ?? []))
      .catch(handleError);
  }
};

export { type Platform, type OfferedVersion, Registry };
