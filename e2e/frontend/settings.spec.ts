import { test, expect } from '@playwright/test';

test.describe('Settings', () => {
  test('settings page renders', async ({ page }) => {
    await page.goto('/#/settings');

    // Should show the settings page content.
    await expect(page.getByText('Authorities')).toBeVisible({ timeout: 10_000 });
  });

  test('displays authorities', async ({ page }) => {
    await page.goto('/#/settings');

    // The bootstrapped authority should be listed.
    await expect(page.getByText('hashicorp')).toBeVisible({ timeout: 10_000 });
  });
});
