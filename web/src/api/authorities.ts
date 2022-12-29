type Key = {
  id: string,
  keyId: string,
  asciiArmor?: string,
  trustSignature?: string,
};

type ApiKey = {
  id: string,
};

type Authority = {
  id: string,
  name: string,
  policyUrl: string,
  keys: Key[],
  apiKeys: ApiKey[]
};

const cache: {
  authorities: Authority[],
} = {
  authorities: [],
};

const fetchAuthorities = (refresh: boolean = false) => {
  if (refresh) {
    ;
  }

  cache.authorities = [
    { 
      id: "1",
      name: "HashiCorp",
      policyUrl: "https://www.hashicorp.com/security",
      keys: [
        { id: "0", keyId: "1" },
        { id: "1", keyId: "2" },
      ], 
      apiKeys: [
        { id: "7290460c-3934-45fd-bb6a-d17dc64b4ae1" },
      ]
    },
    { 
      id: "2", 
      name: "Company", 
      policyUrl: "", 
      keys: [], 
      apiKeys: [] 
    },
  ];

  return cache.authorities;
};

export type { Authority, Key, ApiKey };

export {
  fetchAuthorities
};
