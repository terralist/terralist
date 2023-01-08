import { successAPIResult, errorAPIResult, type APIResult } from "./api.utils";

type Key = {
  id?: string,
  keyId: string,
  asciiArmor?: string,
  trustSignature?: string,
};

type ApiKey = {
  id: string,
};

type Authority = {
  id?: string,
  name: string,
  policyUrl: string,
  keys?: Key[],
  apiKeys?: ApiKey[]
};

const cache: { authorities: Authority[] } = {
  authorities: [],
};

const fetchAuthorities = (refresh: boolean = false): APIResult<Authority[]> => {
  if (refresh) {
    ;
  }

  if (false) {
    return errorAPIResult(500);
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

  return successAPIResult(cache.authorities);
};

const createAuthority = (name: string, policyUrl: string = ""): APIResult<Authority> => {
  let authority = {
    name: name,
    policyUrl: policyUrl,
  } satisfies Authority;

  // TODO: Call create API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(authority);
};

const updateAuthority = (authority: Authority): APIResult<Authority> => {
  // TODO: Call update API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(null);
};

const deleteAuthority = (authority: Authority): APIResult<boolean> => {
  // TODO: Call delete API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(true);
};

const createKey = (authority: Authority, keyId: string, asciiArmor?: string, trustSignature?: string): APIResult<Key> => {
  let key = {
    keyId: keyId,
    asciiArmor: asciiArmor,
    trustSignature: trustSignature,
  } satisfies Key;

  // TODO: Call create authority key API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(key);
};

const updateKey = (authority: Authority, key: Key): APIResult<Key> => {
  // TODO: Call update authority key API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(null);
};

const deleteKey = (authority: Authority, key: Key): APIResult<boolean> => {
  // TODO: Call delete authority key API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(true);
};

const createApiKey = (authority: Authority): APIResult<ApiKey> => {
  // TODO: Call create authority api key API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(null);
};

const deleteApiKey = (authority: Authority, apiKey: ApiKey): APIResult<boolean> => {
  // TODO: Call delete authority key API

  if (false) {
    return errorAPIResult(500);
  }

  return successAPIResult(true);
};

export {
  type Authority,
  fetchAuthorities,
  createAuthority,
  updateAuthority,
  deleteAuthority,

  type Key,
  createKey,
  updateKey,
  deleteKey,

  type ApiKey,
  createApiKey,
  deleteApiKey,
};
