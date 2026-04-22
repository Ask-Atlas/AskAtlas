import { test, expect } from "@playwright/test";

// E2E coverage for DELETE /api/study-guides/{study_guide_id}/recommendations
// (ASK-101).
//
// The endpoint is already shipped -- RemoveRecommendation in
// api/internal/studyguides/service_write.go, handler in
// studyguides.go, with 6 service tests + 5 handler tests covering
// the 204/403/404 paths. This file backfills the missing
// wire-contract verification against dev/staging for 401/400/404.
//
// **Strictly non-destructive coverage**: every test DELETEs against
// the NIL UUID or an invalid UUID string, so no real
// `study_guide_recommendations` row can ever be removed by this
// suite. The 204 happy-path + 403 role-gate path are covered by the
// Go service tests, which use mocked repo expectations and don't
// touch staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("RemoveRecommendation validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(
      `/api/study-guides/${NIL_UUID}/recommendations`,
    );
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid study_guide_id with 400", async ({ request }) => {
    const resp = await request.delete(
      "/api/study-guides/not-a-uuid/recommendations",
    );
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent guide", async ({ request }) => {
    // NIL_UUID never matches a real guide. The service's existence
    // gate returns 404 before any role check or delete fires.
    // Safe to repeat indefinitely against staging.
    const resp = await request.delete(
      `/api/study-guides/${NIL_UUID}/recommendations`,
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});
