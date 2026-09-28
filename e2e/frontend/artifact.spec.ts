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
