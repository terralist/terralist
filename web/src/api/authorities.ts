import { AxiosError } from "axios";
import { createClient, handleResponse, handleError } from "@/api/api.utils";
import type { Key } from "@/api/keys";
import type { ApiKey } from "@/api/apiKeys";

interface Authority {
  id?: string,
  name: string,
  policyUrl: string,
  keys?: Key[],
  apiKeys?: ApiKey[]
};

const client = createClient({
  baseURL: "/v1/api/authorities",
  timeout: 120000,
});

const actions = {
  getAll: () => client
    .get<Authority[]>("/")
    .then(handleResponse<Authority[]>)
    .catch(handleError),

  getOne: (id: string) => client
    .get<Authority>(`/${id}`)
    .then(handleResponse<Authority>)
    .catch(handleError),

  create: (name: string, policyUrl: string) => client
    .post<Authority>("/", { name, policyUrl })
    .then(handleResponse<Authority>)
    .catch(handleError),

  update: (authority: Authority) => {
      if (!authority.id) {
        return Promise.reject(handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, "400")))
      }

      return client
        .patch<Authority>(`/${authority.id}`, authority)
        .then(handleResponse<Authority>)
        .catch(handleError);
    },

  delete: (id: string) => {
    if (!id) {
      return Promise.reject(handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, "400")))
    }

    return client
      .delete<boolean>(`/${id}`)
      .then(handleResponse<boolean>)
      .catch(handleError)
  }
};

const Authorities = {
  getAll: async () => await actions.getAll(),
  getOne: async (id: string) => await actions.getOne(id),
  create: async (name: string, policyUrl: string = "") => await actions.create(name, policyUrl),
  update: async (authority: Authority) => await actions.update(authority),
  delete: async (id: string) => await actions.delete(id),
};

export {
  type Authority,
  type Key,
  type ApiKey,
  Authorities,
};