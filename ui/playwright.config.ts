import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  testMatch: '**/*.spec.ts',
  timeout: 900000, // 15 minutes - Nix builds can take a while
  expect: {
    timeout: 120000, // 2 minutes - first run cache warming can be slow
  },
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    trace: 'on-first-retry',
    video: 'retain-on-failure',
  },
})
