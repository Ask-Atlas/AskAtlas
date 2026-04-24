import { test, expect } from "@playwright/test";

// GET /api/files/{file_id}/download (ASK-205). Non-destructive:
// every test uses NIL_UUID or an invalid UUID so no real bytes are served.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("DownloadFile validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get(`/api/files/${NIL_UUID}/download`, {
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
    const resp = await request.get(`/api/files/${NIL_UUID}/download`, {
      maxRedirects: 0,
    });
    expect(resp.status()).toBe(404);
    expect(resp.headers()["location"]).toBeUndefined();
  });
});
