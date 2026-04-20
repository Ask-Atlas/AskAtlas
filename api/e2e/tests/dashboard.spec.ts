import { test, expect } from "@playwright/test";

// E2E coverage for GET /api/me/dashboard (ASK-155).
//
// The endpoint aggregates 4 sections (courses, study_guides,
// practice, files) in one response. The Go service tests cover
// the term-resolver waterfall, accuracy math, and per-section
// error propagation; this spec pins the wire contract against
// dev/staging:
//   - 401 when unauthenticated
//   - 200 with the full envelope (all 4 sections always present)
//   - List fields always render as [] (not null) when empty
//   - current_term is required+nullable
//   - Numeric fields are integers (not strings)
//   - Per-section field shapes match the openapi spec
//
// Happy-path tests handle "no data" gracefully via shape-only
// assertions (the test user may or may not have any data on
// staging at any given time).

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

test.describe("ListDashboard validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get("/api/me/dashboard");
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("ignores extra query params (no params expected)", async ({
    request,
  }) => {
    // The endpoint takes no query params per spec. Extra params
    // must be silently ignored, not 400'd, so the frontend can
    // append cache-busters etc.
    const resp = await request.get("/api/me/dashboard", {
      params: { foo: "bar", limit: 10 },
    });
    expect(resp.status()).toBe(200);
  });
});

test.describe("ListDashboard envelope shape", () => {
  test("returns 200 with all 4 sections present", async ({ request }) => {
    const resp = await request.get("/api/me/dashboard");
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    expect(body).toHaveProperty("courses");
    expect(body).toHaveProperty("study_guides");
    expect(body).toHaveProperty("practice");
    expect(body).toHaveProperty("files");
  });

  test("courses section has required fields with correct types", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/dashboard");
    expect(resp.status()).toBe(200);
    const c = (await resp.json()).courses;

    expect(c).toHaveProperty("enrolled_count");
    expect(typeof c.enrolled_count).toBe("number");
    expect(Number.isInteger(c.enrolled_count)).toBeTruthy();
    expect(c.enrolled_count).toBeGreaterThanOrEqual(0);

    expect(c).toHaveProperty("current_term");
    // current_term is required+nullable per the openapi schema.
    if (c.current_term !== null) {
      expect(typeof c.current_term).toBe("string");
    }

    expect(c).toHaveProperty("courses");
    expect(Array.isArray(c.courses)).toBeTruthy();
    expect(c.courses.length).toBeLessThanOrEqual(10); // MaxCourses
  });

  test("study_guides section has required fields with correct types", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/dashboard");
    expect(resp.status()).toBe(200);
    const sg = (await resp.json()).study_guides;

    expect(typeof sg.created_count).toBe("number");
    expect(Number.isInteger(sg.created_count)).toBeTruthy();
    expect(sg.created_count).toBeGreaterThanOrEqual(0);

    expect(Array.isArray(sg.recent)).toBeTruthy();
    expect(sg.recent.length).toBeLessThanOrEqual(5); // RecentGuidesLimit
  });

  test("practice section has required fields with correct types", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/dashboard");
    expect(resp.status()).toBe(200);
    const p = (await resp.json()).practice;

    expect(Number.isInteger(p.sessions_completed)).toBeTruthy();
    expect(Number.isInteger(p.total_questions_answered)).toBeTruthy();
    expect(Number.isInteger(p.overall_accuracy)).toBeTruthy();
    // Accuracy is bounded [0, 100].
    expect(p.overall_accuracy).toBeGreaterThanOrEqual(0);
    expect(p.overall_accuracy).toBeLessThanOrEqual(100);

    expect(Array.isArray(p.recent_sessions)).toBeTruthy();
    expect(p.recent_sessions.length).toBeLessThanOrEqual(5); // RecentSessionsLimit
  });

  test("files section has required fields with correct types", async ({
    request,
  }) => {
    const resp = await request.get("/api/me/dashboard");
    expect(resp.status()).toBe(200);
    const f = (await resp.json()).files;

    expect(Number.isInteger(f.total_count)).toBeTruthy();
    expect(f.total_count).toBeGreaterThanOrEqual(0);
    // total_size is bytes, can be a large number.
    expect(typeof f.total_size).toBe("number");
    expect(f.total_size).toBeGreaterThanOrEqual(0);

    expect(Array.isArray(f.recent)).toBeTruthy();
    expect(f.recent.length).toBeLessThanOrEqual(5); // RecentFilesLimit
  });
});

test.describe("ListDashboard per-row shapes", () => {
  test("each course row matches DashboardCourseSummary schema", async ({
    request,
  }) => {
    const body = await (await request.get("/api/me/dashboard")).json();
    if (body.courses.courses.length === 0) {
      test.skip(true, "Test user has no enrollments on this env");
      return;
    }
    for (const c of body.courses.courses) {
      expect(c.id).toMatch(UUID_PATTERN);
      expect(typeof c.department).toBe("string");
      expect(typeof c.number).toBe("string");
      expect(typeof c.title).toBe("string");
      expect(["student", "ta", "instructor"]).toContain(c.role);
      expect(typeof c.section_term).toBe("string");
    }
  });

  test("each study_guide row matches DashboardStudyGuideSummary schema", async ({
    request,
  }) => {
    const body = await (await request.get("/api/me/dashboard")).json();
    if (body.study_guides.recent.length === 0) {
      test.skip(true, "Test user has no created study guides on this env");
      return;
    }
    for (const g of body.study_guides.recent) {
      expect(g.id).toMatch(UUID_PATTERN);
      expect(typeof g.title).toBe("string");
      expect(typeof g.course_department).toBe("string");
      expect(typeof g.course_number).toBe("string");
      expect(Number.isNaN(Date.parse(g.updated_at))).toBeFalsy();
    }
  });

  test("each session row matches DashboardSessionSummary schema", async ({
    request,
  }) => {
    const body = await (await request.get("/api/me/dashboard")).json();
    if (body.practice.recent_sessions.length === 0) {
      test.skip(true, "Test user has no completed sessions on this env");
      return;
    }
    for (const s of body.practice.recent_sessions) {
      expect(s.id).toMatch(UUID_PATTERN);
      expect(typeof s.quiz_title).toBe("string");
      expect(typeof s.study_guide_title).toBe("string");
      expect(Number.isInteger(s.score_percentage)).toBeTruthy();
      expect(s.score_percentage).toBeGreaterThanOrEqual(0);
      expect(s.score_percentage).toBeLessThanOrEqual(100);
      expect(Number.isNaN(Date.parse(s.completed_at))).toBeFalsy();
    }
  });

  test("each file row matches DashboardFileSummary schema", async ({
    request,
  }) => {
    const body = await (await request.get("/api/me/dashboard")).json();
    if (body.files.recent.length === 0) {
      test.skip(true, "Test user has no files on this env");
      return;
    }
    for (const f of body.files.recent) {
      expect(f.id).toMatch(UUID_PATTERN);
      expect(typeof f.name).toBe("string");
      expect(typeof f.mime_type).toBe("string");
      expect(Number.isNaN(Date.parse(f.updated_at))).toBeFalsy();
    }
  });
});
