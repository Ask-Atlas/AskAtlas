import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "html",
  use: {
    baseURL: process.env.E2E_BASE_URL || "https://api-dev.askatlas.study",
    extraHTTPHeaders: {
      Authorization: `Bearer ${process.env.E2E_TOKEN}`,
    },
    trace: "on-first-retry",
  },
});
