import { AxiosError } from 'axios';
import { createClient, handleResponse, handleError } from '@/api/api.utils';

type RuleKind = 'provider' | 'module';
type RuleEffect = 'allow' | 'deny';

type Rule = {
  id: string;
  kind: RuleKind;
  name: string;
  version: string;
  effect: RuleEffect;
};

type CreateRule = Omit<Rule, 'id'>;

const client = createClient({
  baseURL: '/v1/api/authorities',
  timeout: 120000
});

const actions = {
  create: async (authorityId: string, rule: CreateRule) =>
    client
      .post<Rule>(`/${authorityId}/rules`, rule)
      .then(handleResponse<Rule>)
      .catch(handleError),

  delete: async (authorityId: string, id: string) => {
    if (!id) {
      return Promise.reject(
        handleError(new AxiosError(AxiosError.ERR_BAD_REQUEST, '400'))
      );
    }

    return client
      .delete<boolean>(`/${authorityId}/rules/${id}`)
      .then(handleResponse<boolean>)
      .catch(handleError);
  }
};

const Rules = {
  create: async (authorityId: string, rule: CreateRule) =>
    await actions.create(authorityId, rule),
  delete: async (authorityId: string, id: string) =>
    await actions.delete(authorityId, id)
};

export { type Rule, type RuleKind, type RuleEffect, type CreateRule, Rules };
