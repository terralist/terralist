import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('login page shows OAuth provider buttons', async ({ page }) => {
    await page.goto('/');

    // Should redirect to login.
    await expect(page).toHaveURL(/.*login.*/);

    // Should show the OIDC provider button.
    await expect(page.getByText('Continue with OIDC')).toBeVisible();
  });

  test('login via OIDC redirects to dashboard', async ({ page }) => {
    await page.goto('/');

    // Click the OIDC login button — the mock server auto-redirects back.
    await page.getByText('Continue with OIDC').click();

    // Should redirect back to the dashboard.
    await page.waitForURL('**/', { timeout: 10_000 });
    await expect(page.locator('body')).not.toContainText('Continue with OIDC');
  });
});
