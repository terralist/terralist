import { AxiosError } from 'axios';
import { createClient, handleError, handleResponse } from '@/api/api.utils';

type PolicyDTO = {
  id: string;
  resource: string;
  action: string;
  object: string;
  effect: string;
};

type ApiKey = {
  id: string;
  name: string;
  scope: string;
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

type CreateApiKeyDTO = {
  name: string;
  scope: string;
  expireIn: number;
  policies: CreatePolicyDTO[];
};

type CreateApiKeyResponse = {
  id: string;
  name: string;
  key: string;
};

const client = createClient({
  baseURL: '/v1/api/api-keys',
  timeout: 120000
});

const actions = {
  list: async () =>
    client
      .get<ApiKey[]>('/')
      .then(handleResponse<ApiKey[]>)
      .catch(handleError),

  create: async (dto: CreateApiKeyDTO) =>
    client
      .post<CreateApiKeyResponse>('/', dto)
      .then(handleResponse<CreateApiKeyResponse>)
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

const ApiKeys = {
  list: async () => await actions.list(),
  create: async (dto: CreateApiKeyDTO) => await actions.create(dto),
  delete: async (id: string) => await actions.delete(id)
};

export {
  type PolicyDTO,
  type ApiKey,
  type CreatePolicyDTO,
  type CreateApiKeyDTO,
  ApiKeys
};
