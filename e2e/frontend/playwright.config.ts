import { defineConfig } from '@playwright/test';
import path from 'path';

const AUTH_STATE_PATH = path.join(__dirname, '.auth-state.json');

export default defineConfig({
  testDir: '.',
  testMatch: '**/*.spec.ts',
  timeout: 30_000,
  retries: 0,
  use: {
    baseURL: process.env.TERRALIST_URL || 'http://localhost:5758',
    headless: true,
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'setup',
      testMatch: 'auth.setup.ts',
    },
    {
      name: 'auth',
      testMatch: 'auth.spec.ts',
    },
    {
      name: 'authenticated',
      testMatch: ['dashboard.spec.ts', 'settings.spec.ts', 'markdown.spec.ts'],
      dependencies: ['setup'],
      use: {
        storageState: AUTH_STATE_PATH,
      },
    },
  ],
});
