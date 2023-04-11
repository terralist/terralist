import axios, { AxiosError } from "axios";
import { handleResponse, handleError } from "@/api/api.utils";

interface Key {
  id?: string,
  keyId: string,
  asciiArmor?: string,
  trustSignature?: string,
}

const client = axios.create({
  baseURL: "/v1/api/authority",
  timeout: 120,
});

const actions = {
  create: (authorityId: string, keyId: string, asciiArmor?: string, trustSignature?: string) => client
    .post<Key>(`/${authorityId}/key`, { keyId, asciiArmor, trustSignature })
    .then(handleResponse<Key>)
    .catch(handleError),

  delete: (authorityId: string, id: string) => {
    if (!id) {
      return Promise.reject(handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, "400")))
    }

    return client
      .delete<boolean>(`/${authorityId}/key/${id}`)
      .then(handleResponse<boolean>)
      .catch(handleError)
  }
};

const Keys = {
  create: async (authorityId: string, keyId: string, asciiArmor?: string, trustSignature?: string) => await actions.create(authorityId, keyId, asciiArmor, trustSignature),
  delete: async (authorityId: string, id: string) => await actions.delete(authorityId, id),
};

export {
  type Key,
  Keys
};