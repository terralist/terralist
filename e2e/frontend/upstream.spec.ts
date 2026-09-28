import { test, expect, type APIRequestContext } from '@playwright/test';

const masterKey =
  process.env.TERRALIST_MASTER_API_KEY ||
  'e2e-master-api-key-00000000-0000-0000-0000-000000000000';

const api = (request: APIRequestContext) => ({
  headers: { Authorization: `Bearer x-api-key:${masterKey}` },

  async createAuthority(name: string, upstream = {}) {
    const resp = await request.post('/v1/api/authorities', {
      headers: this.headers,
      data: { name, policy_url: '', ...upstream }
    });
    expect(resp.ok()).toBeTruthy();

    return (await resp.json()) as { id: string };
  },

  async addKey(id: string, keyId: string) {
    const resp = await request.post(`/v1/api/authorities/${id}/keys`, {
      headers: this.headers,
      data: { key_id: keyId, ascii_armor: 'armor', trust_signature: '' }
    });
    expect(resp.ok()).toBeTruthy();
  },

  async getAuthority(id: string) {
    const resp = await request.get(`/v1/api/authorities/${id}`, {
      headers: this.headers
    });
    expect(resp.ok()).toBeTruthy();

    return resp.json();
  }
});

test.describe('Authority upstream', () => {
  test('stands for an upstream namespace and pulls through', async ({
    page,
    request
  }) => {
    const name = `ui-upstream-${Date.now()}`;
    const { id } = await api(request).createAuthority(name);
    await api(request).addKey(id, 'KEEPME');

    await page.goto('/#/settings');
    const row = page.getByTestId(`authority-${name}`);
    await row.getByRole('button', { name: 'Edit authority' }).click();

    await page.locator('#upstreamHostname').fill('registry.terraform.io');
    await page.locator('#upstreamNamespace').fill(name);
    await page.locator('#upstreamToken').fill('upstream-token');
    await page.locator('#upstreamPolicy').selectOption('deny');
    await page.getByTestId('checkbox-upstreamEnabled').click();
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(row).toContainText(`registry.terraform.io/${name}`);

    const stored = await api(request).getAuthority(id);
    expect(stored.upstream_hostname).toBe('registry.terraform.io');
    expect(stored.upstream_namespace).toBe(name);
    expect(stored.upstream_enabled).toBe(true);
    expect(stored.upstream_default_policy).toBe('deny');
    expect(stored.upstream_has_token).toBe(true);
    // An update replaces the whole authority, so the keys must be sent back.
    expect(stored.keys.map((k: { key_id: string }) => k.key_id)).toEqual(['KEEPME']);

    // The stored token is never shown, and keeping the field empty keeps it.
    await row.getByRole('button', { name: 'Edit authority' }).click();
    await expect(page.locator('#upstreamToken')).toHaveValue('');
    await expect(page.locator('#upstreamToken')).toHaveAttribute(
      'placeholder',
      /stored/
    );
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect
      .poll(async () => (await api(request).getAuthority(id)).upstream_has_token)
      .toBe(true);
  });
});

test.describe('Upstream rules', () => {
  test('are added and removed from the settings page', async ({
    page,
    request
  }) => {
    const name = `ui-rules-${Date.now()}`;
    const { id } = await api(request).createAuthority(name, {
      upstream_hostname: 'registry.terraform.io',
      upstream_enabled: true
    });

    await page.goto('/#/settings');
    const row = page.getByTestId(`authority-${name}`);
    await row.getByRole('button', { name: 'Add upstream rule' }).click();

    await page.locator('#ruleKind').selectOption('provider');
    await page.locator('#ruleName').fill('aws');
    await page.locator('#ruleVersion').fill('5.*');
    await page.locator('#ruleEffect').selectOption('deny');
    await page.getByRole('button', { name: 'Continue' }).click();

    await row.getByRole('button', { name: 'Show upstream rules' }).click();
    const rule = row.getByTestId('rule-provider-aws-5.*');
    await expect(rule).toContainText('deny');

    const stored = await api(request).getAuthority(id);
    expect(stored.rules).toMatchObject([
      { kind: 'provider', name: 'aws', version: '5.*', effect: 'deny' }
    ]);

    await rule.getByRole('button', { name: 'Remove rule' }).click();
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(rule).toHaveCount(0);
    await expect
      .poll(async () => (await api(request).getAuthority(id)).rules)
      .toEqual([]);
  });
});

test.describe('Upstream rules on a phone', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('can be added from the authority row', async ({ page, request }) => {
    const name = `ui-rules-phone-${Date.now()}`;
    await api(request).createAuthority(name, {
      upstream_hostname: 'registry.terraform.io',
      upstream_enabled: true
    });

    await page.goto('/#/settings');
    const row = page.getByTestId(`authority-${name}`);

    await expect(
      row.getByRole('button', { name: 'Add upstream rule' })
    ).toBeVisible();
  });
});
