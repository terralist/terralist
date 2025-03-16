import { AxiosError } from 'axios';
import { createClient, handleError, handleResponse } from '@/api/api.utils';

type ApiKey = {
  id: string;
  name: string;
};

const client = createClient({
  baseURL: '/v1/api/authorities',
  timeout: 120000
});

const actions = {
  create: async (authorityId: string, name: string) =>
    client
      .post<ApiKey>(`/${authorityId}/api-keys`, { name })
      .then(handleResponse<ApiKey>)
      .catch(handleError),

  delete: async (authorityId: string, id: string) => {
    if (!id) {
      return Promise.reject(
        handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, '400'))
      );
    }

    return client
      .delete<boolean>(`/${authorityId}/api-keys/${id}`)
      .then(handleResponse<boolean>)
      .catch(handleError);
  }
};

const ApiKeys = {
  create: async (authorityId: string, name: string) =>
    await actions.create(authorityId, name),
  delete: async (authorityId: string, id: string) =>
    await actions.delete(authorityId, id)
};

export { type ApiKey, ApiKeys };
