interface Artifact {
  type: "module" | "provider",
  namespace: string,
  name: string,
  provider?: string | undefined,
  version: string
};

const pathRegex = /^\/(?<namespace>[^/]+)\/(?<name>[^/]+)(?:\/(?<provider>[^/]+))?\/(?<version>[^/]+)$/gm;

const fromPath = (path: string) => {
  let m = pathRegex.exec(path);

  if (m === null) {
    return null;
  }

  return {
    type: m.groups.provider ? "module" : "provider",
    namespace: m.groups.namespace,
    name: m.groups.name,
    provider: m.groups.provider,
    version: m.groups.version,
  } satisfies Artifact;
};

export {
  fromPath,
  type Artifact
};