import axios, { AxiosError } from "axios";
import { createClient, handleError, handleResponse } from "@/api/api.utils";

interface ApiKey {
  id: string,
  label: string,
}

const client = createClient({
  baseURL: "/v1/api/authority",
  timeout: 120,
});

const actions = {
  create: (authorityId: string, label: string) => client
    .post<ApiKey>(`/${authorityId}/api-key`, { label })
    .then(handleResponse<ApiKey>)
    .catch(handleError),

  delete: (authorityId: string, id: string) => {
    if (!id) {
      return Promise.reject(handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, "400")))
    }

    return client
      .delete<boolean>(`/${authorityId}/api-key/${id}`)
      .then(handleResponse<boolean>)
      .catch(handleError)
  }
};

const ApiKeys = {
  create: async (authorityId: string, label: string) => await actions.create(authorityId, label),
  delete: async (authorityId: string, id: string) => await actions.delete(authorityId, id),
};

export {
  type ApiKey,
  ApiKeys
};