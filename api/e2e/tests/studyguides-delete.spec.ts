import { test, expect } from "@playwright/test";

// E2E coverage for DELETE /api/study-guides/{study_guide_id} (ASK-133).
//
// The endpoint was already shipped in PR #138 with Go service tests
// covering all 6 outcome paths (success, 404 missing, 404 already-
// deleted, 403 not-creator, lock error 500, quiz cascade error 500).
// This file backfills the missing wire-contract verification against
// dev/staging for the 401/400/404 paths.
//
// **Strictly non-destructive coverage**: every test calls DELETE
// against the NIL UUID or an invalid string, so no real guide can
// ever be soft-deleted (and no quizzes cascaded) by this suite.
// The 204 happy-path + 403 not-creator path are covered by the Go
// service tests, which use mocked repo expectations and don't risk
// staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("DeleteStudyGuide validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(`/api/study-guides/${NIL_UUID}`);
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid guide UUID with 400", async ({ request }) => {
    const resp = await request.delete("/api/study-guides/not-a-uuid");
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for a non-existent guide", async ({ request }) => {
    const resp = await request.delete(`/api/study-guides/${NIL_UUID}`);
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});
