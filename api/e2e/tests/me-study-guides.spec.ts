import { test, expect } from "@playwright/test";

// E2E coverage for GET /api/me/study-guides (ASK-131).
//
// Service-layer tests in api/internal/studyguides/service_test.go
// cover the full matrix (3 sort variants / course_id filter /
// has_more + cursor round-trip / unknown sort 400 / limit clamp /
// soft-deleted rows surfaced). This file backfills the missing
// wire-contract verification against dev/staging for the 401 and
// 400 paths.
//
// **Inherently non-destructive**: GET never mutates state. The
// 200 response contents depend on the authenticated user's own
// guides, so we avoid asserting on the payload shape beyond
// "has expected envelope keys" -- the Go handler tests carry the
// behavioral assertions.

test.describe("ListMyStudyGuides validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get("/api/me/study-guides");
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid sort_by with 400", async ({ request }) => {
    const resp = await request.get(
      "/api/me/study-guides?sort_by=popularity",
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects limit below 1 with 400", async ({ request }) => {
    const resp = await request.get("/api/me/study-guides?limit=0");
    expect(resp.status()).toBe(400);
  });

  test("rejects limit above 100 with 400", async ({ request }) => {
    const resp = await request.get("/api/me/study-guides?limit=101");
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid course_id UUID with 400", async ({ request }) => {
    const resp = await request.get(
      "/api/me/study-guides?course_id=not-a-uuid",
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects malformed cursor with 400", async ({ request }) => {
    const resp = await request.get(
      "/api/me/study-guides?cursor=not-valid-base64!!",
    );
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 400);
    expect(body.details?.cursor).toMatch(/invalid cursor/i);
  });

  test("returns 200 with paginated envelope for authenticated user", async ({
    request,
  }) => {
    // GET is non-destructive; the viewer may have 0 or many guides.
    // Assert only on envelope shape so the test is stable across
    // dev/staging data states.
    const resp = await request.get("/api/me/study-guides");
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body).toHaveProperty("study_guides");
    expect(Array.isArray(body.study_guides)).toBe(true);
    expect(body).toHaveProperty("has_more");
    expect(typeof body.has_more).toBe("boolean");
    expect(body).toHaveProperty("next_cursor");

    // Schema contract: deleted_at is required + nullable on every
    // returned row, so it must be PRESENT (even if null) for
    // soft-delete-aware frontends.
    if (body.study_guides.length > 0) {
      for (const g of body.study_guides) {
        expect(g).toHaveProperty("deleted_at");
      }
    }
  });
});
