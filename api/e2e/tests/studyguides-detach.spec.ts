import { test, expect } from "@playwright/test";

// E2E coverage for the study-guide detach endpoints (ASK-124 + ASK-116):
//
//   DELETE /api/study-guides/{study_guide_id}/files/{file_id}        -- ASK-124
//   DELETE /api/study-guides/{study_guide_id}/resources/{resource_id} -- ASK-116
//
// Both endpoints were already shipped (DetachFile + DetachResource
// services + handlers, with full Go test coverage of the
// dual-authorization branches: guide-creator-OR-file-owner for
// DetachFile, guide-creator-OR-attached-by for DetachResource). This
// file backfills the missing wire-contract verification against
// dev/staging for the 401/400/404 paths.
//
// **Strictly non-destructive coverage**: every test targets the NIL
// UUID for both path params, so no real `study_guide_files` or
// `study_guide_resources` join row can ever be removed by this
// suite. The 204 happy-path + 403 not-authorized path are covered
// by the Go service tests, which use mocked repo expectations and
// don't touch staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("DetachFile validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(
      `/api/study-guides/${NIL_UUID}/files/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid study_guide_id with 400", async ({ request }) => {
    const resp = await request.delete(
      `/api/study-guides/not-a-uuid/files/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid file_id with 400", async ({ request }) => {
    const resp = await request.delete(
      `/api/study-guides/${NIL_UUID}/files/not-a-uuid`,
    );
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent guide", async ({ request }) => {
    // NIL_UUID never matches a real guide, so the locked SELECT
    // returns sql.ErrNoRows -> 404 before any join lookup runs.
    const resp = await request.delete(
      `/api/study-guides/${NIL_UUID}/files/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});

test.describe("DetachResource validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(
      `/api/study-guides/${NIL_UUID}/resources/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid study_guide_id with 400", async ({ request }) => {
    const resp = await request.delete(
      `/api/study-guides/not-a-uuid/resources/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid resource_id with 400", async ({ request }) => {
    const resp = await request.delete(
      `/api/study-guides/${NIL_UUID}/resources/not-a-uuid`,
    );
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent guide", async ({ request }) => {
    const resp = await request.delete(
      `/api/study-guides/${NIL_UUID}/resources/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});
