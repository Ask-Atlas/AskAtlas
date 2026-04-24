import { test, expect } from "@playwright/test";

// E2E coverage for GET /api/files/{file_id}/download (ASK-205).
//
// Service-layer tests in api/internal/files/service_test.go cover the
// happy path (owner + granted-view -> 302 with presigned URL) and all
// state-gating branches (pending/failed/deleted -> 404) against mocked
// repo + generator expectations. Handler-layer tests in
// api/internal/handlers/files_test.go pin the 302 wire shape
// (Location + Cache-Control: no-store + empty body).
//
// This file backfills the missing wire-contract verification against
// dev/staging for the 401 / 400 / 404 paths using the repo's
// non-destructive convention: every test hits the NIL UUID or an
// invalid UUID string, so no real file bytes are ever served and no
// real S3 object is ever touched.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("DownloadFile validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get(`/api/files/${NIL_UUID}/download`, {
      // Don't let Playwright eat the redirect -- we want to verify
      // the 302 status and Location header directly if we ever hit
      // it, and prevent chasing the request into S3 unauthenticated.
      maxRedirects: 0,
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid file UUID with 400", async ({ request }) => {
    const resp = await request.get("/api/files/not-a-uuid/download", {
      maxRedirects: 0,
    });
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for a non-existent file", async ({ request }) => {
    // NIL_UUID is reserved and never assigned to a real file row, so
    // the GetFileIfViewable probe always returns zero rows and the
    // service collapses missing/no-grant/soft-deleted all to 404.
    // Safe to repeat indefinitely against staging.
    const resp = await request.get(`/api/files/${NIL_UUID}/download`, {
      maxRedirects: 0,
    });
    expect(resp.status()).toBe(404);
    expect(resp.headers()["location"]).toBeUndefined();
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });

  test("404 response does not set a Location header", async ({ request }) => {
    // Defense-in-depth: confirms the handler doesn't accidentally
    // write Location before deciding the file isn't servable. A stray
    // Location on a 4xx would leak empty / stale URLs to observers.
    const resp = await request.get(`/api/files/${NIL_UUID}/download`, {
      maxRedirects: 0,
    });
    expect(resp.status()).toBe(404);
    expect(resp.headers()["location"]).toBeUndefined();
  });
});
