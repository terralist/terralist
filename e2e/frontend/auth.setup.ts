import { test as setup, expect } from '@playwright/test';
import path from 'path';

export const AUTH_STATE_PATH = path.join(__dirname, '.auth-state.json');

setup('authenticate via OIDC', async ({ page }) => {
  await page.goto('/');

  // Click the OIDC login button — the mock server auto-redirects back.
  await page.getByText('Continue with OIDC').click();

  // Wait for redirect back to the dashboard.
  await page.waitForURL('**/', { timeout: 10_000 });
  await expect(page.locator('body')).not.toContainText('Continue with OIDC');

  // Save the authenticated state for reuse.
  await page.context().storageState({ path: AUTH_STATE_PATH });
});
