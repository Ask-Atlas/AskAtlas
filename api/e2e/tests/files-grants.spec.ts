import { test, expect } from "@playwright/test";

// E2E coverage for file-grant endpoints (ASK-122 / ASK-125):
//
//   POST   /api/files/{file_id}/grants  -- ASK-122 (create grant)
//   DELETE /api/files/{file_id}/grants  -- ASK-125 (revoke grant)
//
// Service-layer tests in api/internal/files/service_grant_test.go
// cover the full matrix (success / 403 not-owner / 404 missing /
// 409 duplicate / 400 grantee not found / public sentinel exemption)
// against mocked repo expectations. This file backfills the missing
// wire-contract verification against dev/staging for the 401/400/404
// paths -- the surface chi + oapi-codegen + the service enum check
// enforce before any DB write fires.
//
// **Strictly non-destructive coverage**: every test targets the NIL
// UUID file_id, so no real file_grants row can ever be inserted or
// deleted by this suite. The 201 happy path + 403 + 409 paths are
// covered by the Go service tests, which use mocked repo
// expectations and don't touch staging data.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

const validBody = {
  grantee_type: "user",
  grantee_id: NIL_UUID, // public sentinel
  permission: "view",
};

test.describe("CreateGrant validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.post(`/api/files/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid file UUID with 400", async ({ request }) => {
    const resp = await request.post("/api/files/not-a-uuid/grants", {
      data: validBody,
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid grantee_type with 400", async ({ request }) => {
    const resp = await request.post(`/api/files/${NIL_UUID}/grants`, {
      data: { ...validBody, grantee_type: "organization" },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.details?.grantee_type).toMatch(/user.*course.*study_guide/);
  });

  test("rejects invalid permission with 400", async ({ request }) => {
    const resp = await request.post(`/api/files/${NIL_UUID}/grants`, {
      data: { ...validBody, permission: "edit" },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.details?.permission).toMatch(/view.*share.*delete/);
  });

  test("returns 404 for non-existent file", async ({ request }) => {
    // Valid body + valid file UUID format, but the file_id is
    // NIL_UUID which is reserved and never assigned to a real file.
    // Service rejects with 404 before any grantee validation runs.
    const resp = await request.post(`/api/files/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});

test.describe("RevokeGrant validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(`/api/files/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid file UUID with 400", async ({ request }) => {
    const resp = await request.delete("/api/files/not-a-uuid/grants", {
      data: validBody,
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid grantee_type with 400", async ({ request }) => {
    const resp = await request.delete(`/api/files/${NIL_UUID}/grants`, {
      data: { ...validBody, grantee_type: "team" },
    });
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent file", async ({ request }) => {
    const resp = await request.delete(`/api/files/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(404);
  });
});
