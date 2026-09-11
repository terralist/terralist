import { describe, expect, it } from 'vitest';

import { buildPolicyObject, emptyPolicyRow, resources } from './policy';

describe('buildPolicyObject', () => {
  it('builds the object for every resource from its structured fields', (): void => {
    const row = {
      ...emptyPolicyRow(),
      authority: 'hashicorp',
      module: 'subnets',
      provider: 'aws',
      apiKey: 'team-a',
      hostname: 'registry.terraform.io',
      namespace: 'hashicorp'
    };

    expect(buildPolicyObject({ ...row, resource: 'modules' })).toBe(
      'hashicorp/subnets/aws'
    );
    expect(buildPolicyObject({ ...row, resource: 'providers' })).toBe(
      'hashicorp/aws'
    );
    expect(buildPolicyObject({ ...row, resource: 'authorities' })).toBe(
      'hashicorp'
    );
    expect(buildPolicyObject({ ...row, resource: 'api-keys' })).toBe('team-a');
    expect(buildPolicyObject({ ...row, resource: 'mirror' })).toBe(
      'registry.terraform.io/hashicorp/aws'
    );
    expect(buildPolicyObject({ ...row, resource: '*' })).toBe('*');
  });

  it('defaults every field to a wildcard', (): void => {
    for (const resource of resources) {
      const object = buildPolicyObject({ ...emptyPolicyRow(), resource });

      expect(object.split('/').every(part => part === '*')).toBe(true);
    }
  });

  it('offers the mirror resource', (): void => {
    expect(resources).toContain('mirror');
  });
});
