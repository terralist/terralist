// RBAC resources an API key policy can target.
export const resources = [
  'modules',
  'providers',
  'authorities',
  'api-keys',
  '*'
];

// A policy as edited in the API key form. The object is kept as structured
// fields and assembled per resource by buildPolicyObject.
export type PolicyRow = {
  resource: string;
  action: string;
  effect: string;
  authority: string;
  module: string;
  provider: string;
  apiKey: string;
};

export function emptyPolicyRow(): PolicyRow {
  return {
    resource: 'modules',
    action: 'get',
    effect: 'allow',
    authority: '*',
    module: '*',
    provider: '*',
    apiKey: '*'
  };
}

// buildPolicyObject assembles the policy object in the syntax expected by the
// server for the policy resource.
export function buildPolicyObject(policy: PolicyRow): string {
  switch (policy.resource) {
    case 'modules':
      return `${policy.authority}/${policy.module}/${policy.provider}`;
    case 'providers':
      return `${policy.authority}/${policy.provider}`;
    case 'authorities':
      return policy.authority;
    case 'api-keys':
      return policy.apiKey;
    case '*':
      return '*';
    default:
      return '*';
  }
}
