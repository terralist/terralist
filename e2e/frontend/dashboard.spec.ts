import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test('renders after authentication', async ({ page }) => {
    await page.goto('/');

    // Should not be on the login page.
    await expect(page).not.toHaveURL(/.*login.*/);
  });

  test('displays modules and providers', async ({ page }) => {
    await page.goto('/');

    // The dashboard should show the bootstrapped artifacts.
    // The module: hashicorp/subnets/cidr
    await expect(page.getByText('subnets')).toBeVisible({ timeout: 10_000 });

    // The provider: hashicorp/null
    await expect(page.getByText('null')).toBeVisible();
  });

  test('can filter to show only modules', async ({ page }) => {
    await page.goto('/');

    // Wait for content to load.
    await expect(page.getByText('subnets')).toBeVisible({ timeout: 10_000 });

    // Find and click the modules filter toggle.
    const providersToggle = page.locator('label', { hasText: 'Providers' });
    if (await providersToggle.isVisible()) {
      await providersToggle.click();

      // Modules should still be visible.
      await expect(page.getByText('subnets')).toBeVisible();

      // Provider should be hidden.
      await expect(page.getByText('null')).not.toBeVisible();

      // Re-enable providers.
      await providersToggle.click();
    }
  });
});
