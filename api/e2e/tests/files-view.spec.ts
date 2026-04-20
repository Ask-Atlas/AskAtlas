import { test, expect } from "@playwright/test";

// E2E coverage for POST /api/files/{file_id}/view (ASK-134).
//
// Service-layer tests in api/internal/files/service_test.go cover
// the success path (existence -> file_views insert -> file_last_viewed
// upsert) and the failure modes (existence 404, insert error, upsert
// error) against mocked repo expectations. This file backfills the
// missing wire-contract verification against dev/staging for the
// 401/400/404 paths.
//
// **Strictly non-destructive coverage**: every test POSTs against
// the NIL UUID or an invalid UUID string, so no real file_views row
// is ever inserted and no real file_last_viewed timestamp is
// updated. The 204 happy path is covered by the Go service tests,
// which use mocked repo expectations and don't touch staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("RecordFileView validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.post(`/api/files/${NIL_UUID}/view`);
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid file UUID with 400", async ({ request }) => {
    const resp = await request.post("/api/files/not-a-uuid/view");
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for a non-existent file", async ({ request }) => {
    // NIL_UUID is reserved and never assigned to a real file row,
    // so the existence probe always fails before any view write
    // runs. Safe to repeat indefinitely against staging.
    const resp = await request.post(`/api/files/${NIL_UUID}/view`);
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});
