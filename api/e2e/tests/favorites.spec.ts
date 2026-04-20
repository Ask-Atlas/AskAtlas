import { test, expect } from "@playwright/test";

// E2E coverage for GET /api/me/favorites (ASK-151).
//
// The endpoint merges the viewer's favorited files, study guides,
// and courses across three *_favorites tables with offset
// pagination. The Go unit + handler tests cover the
// merge/sort/cursor/filter logic; this spec pins the wire contract
// against dev/staging:
//   - 401 when unauthenticated
//   - 400 on out-of-range limit (0, 101)
//   - 400 on invalid entity_type or cursor
//   - 200 with the documented response shape
//   - Default limit (25), entity_type filter behavior
//   - Cursor opacity (round-trip preserves identity)
//
// Happy-path tests handle "no favorites" gracefully via shape-only
// assertions (the test user may or may not have any favorites on
// staging at any given time).

test.describe("ListFavorites validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get("/api/me/favorites");
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects limit=0 with 400", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { limit: 0 },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects limit=101 with 400", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { limit: 101 },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid entity_type with 400", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { entity_type: "quiz" },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects malformed cursor with 400", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { cursor: "!!!notbase64!!!" },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body).toHaveProperty("details");
    expect(body.details).toHaveProperty("cursor");
  });

  test("accepts boundary limit=1", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { limit: 1 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.favorites)).toBeTruthy();
    expect(body.favorites.length).toBeLessThanOrEqual(1);
  });

  test("accepts boundary limit=100", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { limit: 100 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.favorites)).toBeTruthy();
    expect(body.favorites.length).toBeLessThanOrEqual(100);
  });
});

test.describe("ListFavorites shape", () => {
  test("returns 200 with the documented envelope shape", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/favorites");
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    // Envelope must have all three required fields. favorites is
    // always an array, never null. next_cursor is required +
    // nullable so the field is always present, may be null.
    expect(body).toHaveProperty("favorites");
    expect(Array.isArray(body.favorites)).toBeTruthy();
    expect(body).toHaveProperty("has_more");
    expect(typeof body.has_more).toBe("boolean");
    expect(body).toHaveProperty("next_cursor");
    if (!body.has_more) {
      expect(body.next_cursor).toBeNull();
    } else {
      expect(typeof body.next_cursor).toBe("string");
    }
  });

  test("default limit is 25 when omitted", async ({ request }) => {
    const resp = await request.get("/api/me/favorites");
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.favorites.length).toBeLessThanOrEqual(25);
  });

  test("each item has the discriminated-union envelope", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { limit: 100 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.favorites.length === 0) {
      test.skip(
        true,
        "Test user has no favorites on this env; cannot validate per-row shape",
      );
      return;
    }

    for (const item of body.favorites) {
      expect(item).toHaveProperty("entity_type");
      expect(item).toHaveProperty("entity_id");
      expect(item).toHaveProperty("favorited_at");
      expect(["file", "study_guide", "course"]).toContain(item.entity_type);
      expect(item.entity_id).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
      );
      expect(Number.isNaN(Date.parse(item.favorited_at))).toBeFalsy();

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

  test("results are sorted by favorited_at DESC", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { limit: 100 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.favorites.length < 2) {
      test.skip(
        true,
        `Need >=2 favorites to validate ordering; have ${body.favorites.length}`,
      );
      return;
    }

    for (let i = 1; i < body.favorites.length; i++) {
      const prev = Date.parse(body.favorites[i - 1].favorited_at);
      const cur = Date.parse(body.favorites[i].favorited_at);
      expect(prev).toBeGreaterThanOrEqual(cur);
    }
  });

  test("entity_type=file returns only file items", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { entity_type: "file", limit: 100 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.favorites.length === 0) {
      test.skip(true, "Test user has no file favorites on this env");
      return;
    }
    for (const item of body.favorites) {
      expect(item.entity_type).toBe("file");
      expect(item.file).toBeDefined();
    }
  });

  test("entity_type=study_guide returns only study_guide items", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { entity_type: "study_guide", limit: 100 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.favorites.length === 0) {
      test.skip(true, "Test user has no study_guide favorites on this env");
      return;
    }
    for (const item of body.favorites) {
      expect(item.entity_type).toBe("study_guide");
      expect(item.study_guide).toBeDefined();
    }
  });

  test("entity_type=course returns only course items", async ({ request }) => {
    const resp = await request.get("/api/me/favorites", {
      params: { entity_type: "course", limit: 100 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    if (body.favorites.length === 0) {
      test.skip(true, "Test user has no course favorites on this env");
      return;
    }
    for (const item of body.favorites) {
      expect(item.entity_type).toBe("course");
      expect(item.course).toBeDefined();
    }
  });

  test("cursor round-trips: page two starts where page one ended", async ({
    request,
  }) => {
    // Use limit=1 to maximize the chance of a multi-page result on
    // a sparsely-populated test account.
    const first = await request.get("/api/me/favorites", {
      params: { limit: 1 },
    });
    expect(first.status()).toBe(200);
    const firstBody = await first.json();

    if (!firstBody.has_more) {
      test.skip(
        true,
        "Test user has 0 or 1 favorites; cannot validate cursor round-trip",
      );
      return;
    }
    expect(typeof firstBody.next_cursor).toBe("string");
    expect(firstBody.favorites.length).toBe(1);

    const second = await request.get("/api/me/favorites", {
      params: { limit: 1, cursor: firstBody.next_cursor },
    });
    expect(second.status()).toBe(200);
    const secondBody = await second.json();

    // Second page may or may not have more, but it must contain a
    // different item than the first page (offset advanced).
    if (secondBody.favorites.length > 0) {
      expect(secondBody.favorites[0].entity_id).not.toBe(
        firstBody.favorites[0].entity_id,
      );
    }
  });
});
