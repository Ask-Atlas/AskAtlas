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

// Live-data shape tests. Fetches real entity IDs via list endpoints
// (read-only, non-destructive) and resolves them, asserting the
// populated summary shape matches the OpenAPI contract. Skips when
// the viewer has no visible entities.
test.describe("ResolveRefs live shape (real entities)", () => {
  test("resolves a real study guide to a populated SgRefSummary", async ({
    request,
  }) => {
    const listResp = await request.get("/api/me/study-guides", {
      params: { page_limit: 1 },
    });
    expect(listResp.ok()).toBeTruthy();
    const list = await listResp.json();
    if (!list.study_guides?.length) {
      test.skip(true, "No study guides visible to viewer");
      return;
    }
    const sg = list.study_guides[0];

    const resp = await request.post("/api/refs/resolve", {
      data: { refs: [{ type: "sg", id: sg.id }] },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    const summary = body.results[`sg:${sg.id}`];
    expect(summary).not.toBeNull();
    expect(summary.type).toBe("sg");
    expect(summary.id).toBe(sg.id);
    expect(typeof summary.title).toBe("string");
    expect(summary.title.length).toBeGreaterThan(0);
    // course info mirrored from the joined courses row
    expect(typeof summary.course?.department).toBe("string");
    expect(typeof summary.course?.number).toBe("string");
    // quiz_count is an int derived from a live subquery
    expect(typeof summary.quiz_count).toBe("number");
    expect(summary.quiz_count).toBeGreaterThanOrEqual(0);
    expect(typeof summary.is_recommended).toBe("boolean");
  });

  test("resolves a real course to a populated CourseRefSummary", async ({
    request,
  }) => {
    const listResp = await request.get("/api/courses", {
      params: { page_limit: 1 },
    });
    expect(listResp.ok()).toBeTruthy();
    const list = await listResp.json();
    if (!list.courses?.length) {
      test.skip(true, "No courses on stage");
      return;
    }
    const course = list.courses[0];

    const resp = await request.post("/api/refs/resolve", {
      data: { refs: [{ type: "course", id: course.id }] },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    const summary = body.results[`course:${course.id}`];
    expect(summary).not.toBeNull();
    expect(summary.type).toBe("course");
    expect(summary.id).toBe(course.id);
    expect(typeof summary.title).toBe("string");
    expect(typeof summary.department).toBe("string");
    expect(typeof summary.number).toBe("string");
    expect(typeof summary.school?.name).toBe("string");
    expect(typeof summary.school?.acronym).toBe("string");
  });

  test("mixed batch (real sg + real course + nil file) returns a coherent map", async ({
    request,
  }) => {
    const [sgResp, courseResp] = await Promise.all([
      request.get("/api/me/study-guides", { params: { page_limit: 1 } }),
      request.get("/api/courses", { params: { page_limit: 1 } }),
    ]);
    expect(sgResp.ok()).toBeTruthy();
    expect(courseResp.ok()).toBeTruthy();
    const sgList = await sgResp.json();
    const courseList = await courseResp.json();
    if (!sgList.study_guides?.length || !courseList.courses?.length) {
      test.skip(true, "Stage lacks a study guide or a course for live resolve");
      return;
    }
    const sgId = sgList.study_guides[0].id;
    const courseId = courseList.courses[0].id;

    const resp = await request.post("/api/refs/resolve", {
      data: {
        refs: [
          { type: "sg", id: sgId },
          { type: "course", id: courseId },
          { type: "file", id: NIL_UUID }, // deliberately missing
        ],
      },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Object.keys(body.results)).toHaveLength(3);
    expect(body.results[`sg:${sgId}`]?.type).toBe("sg");
    expect(body.results[`course:${courseId}`]?.type).toBe("course");
    expect(body.results[`file:${NIL_UUID}`]).toBeNull();
  });
});
