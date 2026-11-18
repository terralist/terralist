import { test, expect } from '@playwright/test';

// The OIDC test user is granted role:admin only through the group:e2e-admins
// mapping in e2e/policy.csv, so these checks fail if the groups claim is not
// propagated from the identity provider into the session.
test.describe('RBAC via OIDC groups', () => {
  test('session carries the groups from the OIDC claim', async ({ request }) => {
    const resp = await request.get('/v1/api/auth/session');
    expect(resp.status()).toBe(200);

    const session = await resp.json();
    expect(session.name).toBe('e2e-test-user');
    expect(session.groups).toEqual(['e2e-admins']);
  });

  test('group grant allows access to settings', async ({ request }) => {
    const resp = await request.get('/v1/api/auth/capabilities/settings');
    expect(resp.status()).toBe(200);

    const capability = await resp.json();
    expect(capability.allowed).toBe(true);
  });
});
