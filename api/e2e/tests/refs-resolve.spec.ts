import { test, expect } from "@playwright/test";

// POST /api/refs/resolve (ASK-208). Non-destructive: all tests use
// NIL_UUID or invalid input, so no real rows are looked up beyond
// existence probes that find nothing.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

test.describe("ResolveRefs validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.post("/api/refs/resolve", {
      data: { refs: [{ type: "sg", id: NIL_UUID }] },
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects empty refs array with 400", async ({ request }) => {
    const resp = await request.post("/api/refs/resolve", {
      data: { refs: [] },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.status).toBe("VALIDATION_ERROR");
  });

  test("rejects more than 50 refs with 400", async ({ request }) => {
    const refs = Array.from({ length: 51 }, () => ({
      type: "sg" as const,
      id: NIL_UUID,
    }));
    const resp = await request.post("/api/refs/resolve", {
      data: { refs },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects unknown ref type with 400", async ({ request }) => {
    const resp = await request.post("/api/refs/resolve", {
      data: { refs: [{ type: "user", id: NIL_UUID }] },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid UUID format with 400", async ({ request }) => {
    const resp = await request.post("/api/refs/resolve", {
      data: { refs: [{ type: "sg", id: "not-a-uuid" }] },
    });
    expect(resp.status()).toBe(400);
  });
});

test.describe("ResolveRefs happy shape", () => {
  test("returns results map with null for unknown refs", async ({ request }) => {
    const resp = await request.post("/api/refs/resolve", {
      data: {
        refs: [
          { type: "sg", id: NIL_UUID },
          { type: "quiz", id: NIL_UUID },
          { type: "file", id: NIL_UUID },
          { type: "course", id: NIL_UUID },
        ],
      },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body).toHaveProperty("results");
    expect(typeof body.results).toBe("object");
    for (const type of ["sg", "quiz", "file", "course"] as const) {
      const key = `${type}:${NIL_UUID}`;
      expect(Object.prototype.hasOwnProperty.call(body.results, key)).toBe(
        true,
      );
      // NIL_UUID is never a live row -> null in all four slots.
      expect(body.results[key]).toBeNull();
    }
  });

  test("dedupes duplicate refs to a single map entry", async ({ request }) => {
    const resp = await request.post("/api/refs/resolve", {
      data: {
        refs: [
          { type: "sg", id: NIL_UUID },
          { type: "sg", id: NIL_UUID },
          { type: "sg", id: NIL_UUID },
        ],
      },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Object.keys(body.results)).toHaveLength(1);
    expect(body.results[`sg:${NIL_UUID}`]).toBeNull();
  });
});
