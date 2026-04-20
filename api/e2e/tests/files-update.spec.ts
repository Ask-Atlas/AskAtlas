import { test, expect } from "@playwright/test";

// E2E coverage for PATCH /api/files/{file_id} (ASK-113).
//
// Service-layer tests in api/internal/files/service_test.go cover all
// 10 acceptance criteria (rename, status transitions, 404, 403, empty
// body, etc.) against mocked repo expectations. This file backfills
// the missing wire-contract verification against dev/staging for the
// 401/400/404 paths -- the surface oapi-codegen + chi enforce before
// the handler runs.
//
// **Strictly non-destructive coverage**: every test PATCHes against
// the NIL UUID or an invalid UUID string, so no real file can be
// renamed or transitioned by this suite. Even when the body is a
// valid PATCH payload, the target row never matches, so the worst
// case the service can do is return 404. The 200 happy-path + 403
// not-owner path are covered by the Go service tests, which use
// mocked repo expectations and don't touch staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("UpdateFile validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.patch(`/api/files/${NIL_UUID}`, {
      data: { name: "ignored.pdf" },
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid file UUID with 400", async ({ request }) => {
    const resp = await request.patch("/api/files/not-a-uuid", {
      data: { name: "ignored.pdf" },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects empty body with 400", async ({ request }) => {
    // The spec requires at least one of {name, status}. An empty
    // object must return 400 with the "At least one field" message.
    const resp = await request.patch(`/api/files/${NIL_UUID}`, {
      data: {},
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 400);
    expect(body.message).toMatch(/at least one field/i);
  });

  test("rejects empty name after trim with 400", async ({ request }) => {
    const resp = await request.patch(`/api/files/${NIL_UUID}`, {
      data: { name: "   " },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.details?.name).toMatch(/empty/i);
  });

  test("rejects status outside enum with 400", async ({ request }) => {
    // Spec disallows transitioning to "pending" (or any value that
    // isn't "complete"/"failed"); the service rejects before SQL.
    const resp = await request.patch(`/api/files/${NIL_UUID}`, {
      data: { status: "pending" },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.details?.status).toMatch(/complete.*failed/);
  });

  test("returns 404 for a non-existent file", async ({ request }) => {
    // Valid body, well-formed UUID, but the row simply doesn't exist
    // -- so we always land on 404 regardless of caller identity.
    // Safe to run repeatedly: NIL_UUID is reserved and never assigned
    // to a real file row.
    const resp = await request.patch(`/api/files/${NIL_UUID}`, {
      data: { name: "ignored.pdf" },
    });
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });

  test("returns 404 for a non-existent file when sending status only", async ({
    request,
  }) => {
    // Same row, status-only body. Confirms the 404 path fires before
    // any transition validation -- the row doesn't exist, so we
    // never check whether `pending -> complete` was the right move.
    const resp = await request.patch(`/api/files/${NIL_UUID}`, {
      data: { status: "complete" },
    });
    expect(resp.status()).toBe(404);
  });
});
