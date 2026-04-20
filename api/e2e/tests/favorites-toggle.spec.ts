import { test, expect } from "@playwright/test";

// E2E coverage for the three favorite-toggle endpoints (ASK-130 /
// ASK-156 / ASK-157):
//
//   POST /api/files/{file_id}/favorite                  -- ASK-130
//   POST /api/me/study-guides/{study_guide_id}/favorite -- ASK-156
//   POST /api/me/courses/{course_id}/favorite           -- ASK-157
//
// Service-layer tests in api/internal/favorites/service_test.go
// cover the favorite/unfavorite/404 paths against mocked repo
// expectations. This file backfills the missing wire-contract
// verification against dev/staging for the 401/400/404 paths.
//
// **Strictly non-destructive coverage**: every test POSTs against
// the NIL UUID or an invalid UUID string, so no real file/guide/
// course is ever favorited. The 200 favorite + 200 unfavorite paths
// are covered by the Go service tests, which use mocked repo
// expectations and don't touch staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

const TARGETS = [
  {
    label: "file",
    path: (id: string) => `/api/files/${id}/favorite`,
  },
  {
    label: "study guide",
    path: (id: string) => `/api/me/study-guides/${id}/favorite`,
  },
  {
    label: "course",
    path: (id: string) => `/api/me/courses/${id}/favorite`,
  },
];

for (const target of TARGETS) {
  test.describe(`Toggle ${target.label} favorite validation`, () => {
    test("rejects unauthenticated with 401", async ({ playwright }) => {
      const noAuth = await playwright.request.newContext({
        baseURL: process.env.E2E_BASE_URL,
        extraHTTPHeaders: {},
      });
      const resp = await noAuth.post(target.path(NIL_UUID));
      expect(resp.status()).toBe(401);
      await noAuth.dispose();
    });

    test("rejects invalid UUID with 400", async ({ request }) => {
      const resp = await request.post(target.path("not-a-uuid"));
      expect(resp.status()).toBe(400);
    });

    test("returns 404 for a non-existent target", async ({ request }) => {
      // NIL_UUID is reserved and never assigned to a real row, so
      // this always lands on the existence-probe 404 path before any
      // toggle SQL runs. Safe to repeat indefinitely against staging.
      const resp = await request.post(target.path(NIL_UUID));
      expect(resp.status()).toBe(404);
      const body = await resp.json();
      expect(body).toHaveProperty("code", 404);
      expect(body.message).toMatch(/not found/i);
    });
  });
}
