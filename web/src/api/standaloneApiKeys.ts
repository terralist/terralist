import { AxiosError } from 'axios';
import { createClient, handleError, handleResponse } from '@/api/api.utils';

type PolicyDTO = {
  id: string;
  resource: string;
  action: string;
  object: string;
  effect: string;
};

type StandaloneApiKey = {
  id: string;
  name: string;
  createdBy: string;
  expiration: string;
  policies: PolicyDTO[];
};

type CreatePolicyDTO = {
  resource: string;
  action: string;
  object: string;
  effect: string;
};

type CreateStandaloneApiKeyDTO = {
  name: string;
  expireIn: number;
  policies: CreatePolicyDTO[];
};

type CreateStandaloneApiKeyResponse = {
  id: string;
  name: string;
};

const client = createClient({
  baseURL: '/v1/api/api-keys',
  timeout: 120000
});

const actions = {
  list: async () =>
    client
      .get<StandaloneApiKey[]>('/')
      .then(handleResponse<StandaloneApiKey[]>)
      .catch(handleError),

  create: async (dto: CreateStandaloneApiKeyDTO) =>
    client
      .post<CreateStandaloneApiKeyResponse>('/', dto)
      .then(handleResponse<CreateStandaloneApiKeyResponse>)
      .catch(handleError),

  delete: async (id: string) => {
    if (!id) {
      return Promise.reject(
        handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, '400'))
      );
    }

    return client
      .delete<boolean>(`/${id}`)
      .then(handleResponse<boolean>)
      .catch(handleError);
  }
};

const StandaloneApiKeys = {
  list: async () => await actions.list(),
  create: async (dto: CreateStandaloneApiKeyDTO) => await actions.create(dto),
  delete: async (id: string) => await actions.delete(id)
};

export {
  type PolicyDTO,
  type StandaloneApiKey,
  type CreatePolicyDTO,
  type CreateStandaloneApiKeyDTO,
  StandaloneApiKeys
};
