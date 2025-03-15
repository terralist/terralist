import { AxiosError } from "axios";
import { createClient, handleResponse, handleError } from "@/api/api.utils";

type Key = {
  id: string;
  keyId: string;
  asciiArmor: string;
  trustSignature: string;
}

const client = createClient({
  baseURL: "/v1/api/authorities",
  timeout: 120000,
});

const actions = {
  create: (authorityId: string, keyId: string, asciiArmor?: string, trustSignature?: string) => client
    .post<Key>(`/${authorityId}/keys`, { keyId, asciiArmor, trustSignature })
    .then(handleResponse<Key>)
    .catch(handleError),

  delete: (authorityId: string, id: string) => {
    if (!id) {
      return Promise.reject(handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, "400")))
    }

    return client
      .delete<boolean>(`/${authorityId}/keys/${id}`)
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
