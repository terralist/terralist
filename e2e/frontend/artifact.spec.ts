import { test, expect, type APIRequestContext } from '@playwright/test';

const masterKey =
  process.env.TERRALIST_MASTER_API_KEY ||
  'e2e-master-api-key-00000000-0000-0000-0000-000000000000';

type VersionDetails = {
  version: string;
  origin: string;
  mirror_only?: boolean;
};

const versionsOf = async (
  request: APIRequestContext,
  slug: string
): Promise<VersionDetails[]> => {
  const resp = await request.get(`/v1/api/artifacts/${slug}/version`, {
    headers: { Authorization: `Bearer x-api-key:${masterKey}` }
  });
  expect(resp.ok()).toBeTruthy();

  return (await resp.json()).versions;
};

test.describe('Artifact versions', () => {
  test('shows a provider version served by the network mirror only', async ({
    page,
    request
  }) => {
    const mirrorOnly = (await versionsOf(request, 'hashicorp/null')).find(
      v => v.mirror_only
    );
    expect(mirrorOnly).toBeDefined();

    await page.goto(`/#/providers/hashicorp/null/${mirrorOnly!.version}`);

    await expect(page.getByTestId('version-badges')).toContainText(
      'network mirror only'
    );
  });

  test('shows a provider version pulled from the upstream', async ({
    page,
    request
  }) => {
    const pulled = (await versionsOf(request, 'hashicorp/random')).find(
      v => v.origin === 'upstream'
    );
    expect(pulled).toBeDefined();

    await page.goto(`/#/providers/hashicorp/random/${pulled!.version}`);

    await expect(page.getByTestId('version-badges')).toContainText(
      'pulled from upstream'
    );
  });

  test('shows no badge on an uploaded, signed provider version', async ({
    page,
    request
  }) => {
    const uploaded = (await versionsOf(request, 'hashicorp/null')).find(
      v => v.origin === 'manual' && !v.mirror_only
    );
    expect(uploaded).toBeDefined();

    await page.goto(`/#/providers/hashicorp/null/${uploaded!.version}`);
    await expect(page.getByText(uploaded!.version).first()).toBeVisible();

    await expect(page.getByTestId('version-badges')).toHaveCount(0);
  });
});

const headers = { Authorization: `Bearer x-api-key:${masterKey}` };

type RegistryVersion = {
  version: string;
  platforms: { os: string; arch: string }[];
};

// notHeld returns an upstream version of hashicorp/random that Terralist does
// not hold, with its platforms.
const notHeld = async (
  request: APIRequestContext,
  skip: string[] = []
): Promise<RegistryVersion> => {
  const held = (await versionsOf(request, 'hashicorp/random')).map(
    v => v.version
  );
  const resp = await request.get('/v1/providers/hashicorp/random/versions', {
    headers
  });
  expect(resp.ok()).toBeTruthy();

  const offered: RegistryVersion[] = (await resp.json()).versions;
  const candidate = offered.find(
    v =>
      !held.includes(v.version) &&
      !skip.includes(v.version) &&
      v.platforms.length > 0
  );
  expect(candidate).toBeDefined();

  return candidate!;
};

const hashicorpAuthority = async (request: APIRequestContext) => {
  const resp = await request.get('/v1/api/authorities/', { headers });
  expect(resp.ok()).toBeTruthy();

  const all: { id: string; name: string }[] = await resp.json();
  return all.find(a => a.name === 'hashicorp')!;
};

test.describe('Artifact version actions', () => {
  test('fetches a version from the upstream, then blocks it', async ({
    page,
    request
  }) => {
    const candidate = await notHeld(request);
    const platform = `${candidate.platforms[0].os}_${candidate.platforms[0].arch}`;
    const held = (await versionsOf(request, 'hashicorp/random'))[0].version;

    await page.goto(`/#/providers/hashicorp/random/${held}`);
    await page.getByRole('button', { name: 'Fetch from upstream' }).click();
    await page.locator('#fetchVersion').selectOption(candidate.version);
    await page.locator('#fetchPlatforms').fill(platform);
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(
      new RegExp(`/providers/hashicorp/random/${candidate.version}$`),
      { timeout: 60_000 }
    );
    await expect(page.getByTestId('version-badges')).toContainText(
      'pulled from upstream'
    );

    await page.getByRole('button', { name: 'Block version' }).click();
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).not.toHaveURL(
      new RegExp(`/providers/hashicorp/random/${candidate.version}$`)
    );
    expect(
      (await versionsOf(request, 'hashicorp/random')).map(v => v.version)
    ).not.toContain(candidate.version);

    const authority = await hashicorpAuthority(request);
    const stored = await (
      await request.get(`/v1/api/authorities/${authority.id}`, { headers })
    ).json();
    const rule = stored.rules.find(
      (r: { name: string; version: string }) =>
        r.name === 'random' && r.version === candidate.version
    );
    expect(rule).toMatchObject({ kind: 'provider', effect: 'deny' });

    // Let later runs pull the version again.
    await request.delete(
      `/v1/api/authorities/${authority.id}/rules/${rule.id}`,
      { headers }
    );
  });

  test('deletes a version', async ({ page, request }) => {
    const candidate = await notHeld(request);
    const platform = `${candidate.platforms[0].os}_${candidate.platforms[0].arch}`;
    const fetched = await request.post(
      `/v1/api/providers/hashicorp/random/${candidate.version}/fetch`,
      { headers, data: { platforms: [platform] }, timeout: 60_000 }
    );
    expect(fetched.ok()).toBeTruthy();

    await page.goto(`/#/providers/hashicorp/random/${candidate.version}`);
    await page.getByRole('button', { name: 'Delete version' }).click();
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).not.toHaveURL(
      new RegExp(`/providers/hashicorp/random/${candidate.version}$`)
    );
    expect(
      (await versionsOf(request, 'hashicorp/random')).map(v => v.version)
    ).not.toContain(candidate.version);
  });
});
