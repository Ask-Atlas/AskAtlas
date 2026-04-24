import { test, expect } from "@playwright/test";

// E2E coverage for GET /api/me/recents (ASK-145).
//
// The endpoint merges the viewer's most recent files, study guides,
// and courses across three *_last_viewed tables. The Go unit +
// handler tests cover the merge/sort/truncate logic; this spec pins
// the wire contract against dev/staging:
//   - 401 when unauthenticated
//   - 400 on out-of-range limit (0, 31)
//   - 200 with the documented response shape (recents always an array)
//   - Default limit (10) when omitted
//   - viewed_at DESC ordering across whatever the test user has viewed
//
// All happy-path tests handle "no view history" gracefully via
// shape-only assertions (the test user may or may not have any
// recent activity on staging at any given time).

test.describe("ListRecents validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get("/api/me/recents");
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects limit=0 with 400", async ({ request }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: 0 },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects limit=31 with 400", async ({ request }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: 31 },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects non-integer limit with 400", async ({ request }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: "abc" },
    });
    expect(resp.status()).toBe(400);
  });

  test("accepts boundary limit=1", async ({ request }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: 1 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.recents)).toBeTruthy();
    expect(body.recents.length).toBeLessThanOrEqual(1);
  });

  test("accepts boundary limit=30", async ({ request }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: 30 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.recents)).toBeTruthy();
    expect(body.recents.length).toBeLessThanOrEqual(30);
  });
});

test.describe("ListRecents shape", () => {
  test("returns 200 with the documented envelope shape", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/recents");
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    // The recents field must always be an array, never null. The Go
    // handler explicitly initializes a non-nil empty slice so the
    // JSON renders as [] for empty results -- this assertion pins
    // that contract.
    expect(body).toHaveProperty("recents");
    expect(Array.isArray(body.recents)).toBeTruthy();
  });

  test("default limit is 10 when omitted", async ({ request }) => {
    const resp = await request.get("/api/me/recents");
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.recents.length).toBeLessThanOrEqual(10);
  });

  test("each item has the discriminated-union envelope", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: 30 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.recents.length === 0) {
      test.skip(
        true,
        "Test user has no view history on this env; cannot validate per-row shape",
      );
      return;
    }

    for (const item of body.recents) {
      expect(item).toHaveProperty("entity_type");
      expect(item).toHaveProperty("entity_id");
      expect(item).toHaveProperty("viewed_at");
      expect(["file", "study_guide", "course"]).toContain(item.entity_type);
      expect(item.entity_id).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
      );
      expect(Number.isNaN(Date.parse(item.viewed_at))).toBeFalsy();

      // Exactly one of file/study_guide/course is populated; the
      // others are absent (not null) per the openapi spec.
      const populated = [item.file, item.study_guide, item.course].filter(
        (v) => v !== undefined,
      );
      expect(populated.length).toBe(1);

      switch (item.entity_type) {
        case "file":
          expect(item.file).toBeDefined();
          expect(item.file.id).toBe(item.entity_id);
          expect(typeof item.file.name).toBe("string");
          expect(typeof item.file.mime_type).toBe("string");
          break;
        case "study_guide":
          expect(item.study_guide).toBeDefined();
          expect(item.study_guide.id).toBe(item.entity_id);
          expect(typeof item.study_guide.title).toBe("string");
          expect(typeof item.study_guide.course_department).toBe("string");
          expect(typeof item.study_guide.course_number).toBe("string");
          break;
        case "course":
          expect(item.course).toBeDefined();
          expect(item.course.id).toBe(item.entity_id);
          expect(typeof item.course.department).toBe("string");
          expect(typeof item.course.number).toBe("string");
          expect(typeof item.course.title).toBe("string");
          break;
      }
    }
  });

  test("results are sorted by viewed_at DESC", async ({ request }) => {
    const resp = await request.get("/api/me/recents", {
      params: { limit: 30 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.recents.length < 2) {
      test.skip(
        true,
        `Need >=2 recent items to validate ordering; have ${body.recents.length}`,
      );
      return;
    }

    // Walk pairs of consecutive rows: each earlier row's
    // viewed_at must be >= the later row's. Asserts the same
    // ordering the Go service produces (DESC by ViewedAt).
    for (let i = 1; i < body.recents.length; i++) {
      const prev = Date.parse(body.recents[i - 1].viewed_at);
      const cur = Date.parse(body.recents[i].viewed_at);
      expect(prev).toBeGreaterThanOrEqual(cur);
    }
  });
});
