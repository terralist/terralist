const urlPattern = /^(?:(?<scheme>http[s]?|ftp|mailto|file|data|irc):\/)?\/?(?:(?<username>[\w]+)[:](?<password>[\w]+)[@])?(?<host>[^:\/\s]+)(?:[:](?<port>[\d]+))?(?<path>(?:(?:\/\w+)*\/)(?:[\w\-\.]+?)?[^#?\s]*)?(?:[?](?<query>[^#\s]+))?(?:(?:[#])(?<fragment>[\w\-]+))?$/gm;

interface URL {
  scheme: string | undefined
  username: string | undefined
  password: string | undefined
  host: string | undefined
  port: string | undefined
  path: string | undefined
  query: string | undefined
  fragment: string | undefined
};

const parseUrl: (url: string) => URL | null = (url: string) => {
  let m = urlPattern.exec(url);

  if (m === null) {
    return null;
  }

  return {
    scheme: m.groups.scheme,
    username: m.groups.username,
    password: m.groups.password,
    host: m.groups.host,
    port: m.groups.port,
    path: m.groups.path,
    query: m.groups.query,
    fragment: m.groups.fragment,
  } satisfies URL;
};

const currentPath = () => {
  const url = parseUrl(window.location.href);

  return url.path;
};

const redirect = (url: string) => {
  window.location.href = url;
};

export {
  parseUrl,
  currentPath,
  redirect,
};