import { test, expect } from "@playwright/test";

// E2E coverage for the courses surface.
//
// Currently only ASK-127 (GET /courses/{course_id}/sections) is
// covered; sibling endpoints already have their own coverage in
// the legacy contracts.spec.ts. As more course endpoints land or
// need staging-contract verification beyond Go unit tests, they
// get added here.
//
// Discovery: validation tests run unconditionally; the substantive
// happy-path tests look up any course on staging via GET /courses
// (the test user has read access to all courses) and skip cleanly
// when no seed data exists.

test.describe("ListCourseSections (ASK-127)", () => {
  // Shared discovery state. Set in beforeAll; null when no seed
  // data was found (substantive tests early-skip).
  let courseId: string | null = null;

  test.beforeAll(async ({ request }) => {
    const coursesResp = await request.get("/api/courses", {
      params: { page_limit: 1 },
    });
    if (!coursesResp.ok()) return;
    const course = (await coursesResp.json())?.courses?.[0];
    if (!course?.id) return;
    courseId = course.id;
  });

  // ---------- Validation (no seed data required) ----------

  test("rejects unauthenticated", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get(
      "/api/courses/00000000-0000-0000-0000-000000000000/sections",
    );
    expect([401, 403]).toContain(resp.status());
    await noAuth.dispose();
  });

  test("rejects invalid course UUID with 400", async ({ request }) => {
    const resp = await request.get("/api/courses/not-a-uuid/sections");
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for a non-existent course", async ({ request }) => {
    const resp = await request.get(
      "/api/courses/00000000-0000-0000-0000-000000000000/sections",
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });

  test("rejects term that exceeds 30 chars with 400", async ({ request }) => {
    const tooLong = "a".repeat(31);
    const resp = await request.get(
      "/api/courses/00000000-0000-0000-0000-000000000000/sections",
      { params: { term: tooLong } },
    );
    expect(resp.status()).toBe(400);
  });

  // ---------- Happy path (requires seed data) ----------

  test("returns 200 with ListCourseSectionsResponse shape", async ({
    request,
  }) => {
    if (!courseId) {
      test.skip(true, "No seed course available on staging");
      return;
    }

    const resp = await request.get(`/api/courses/${courseId}/sections`);
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();

    expect(body).toHaveProperty("sections");
    expect(Array.isArray(body.sections)).toBeTruthy();

    if (body.sections.length === 0) {
      test.skip(
        true,
        `Seed course ${courseId} has no sections; cannot validate per-row shape`,
      );
      return;
    }

    const s = body.sections[0];
    expect(s.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
    expect(s.course_id).toBe(courseId);
    expect(typeof s.term).toBe("string");
    expect(s.term.length).toBeGreaterThan(0);
    // section_code + instructor_name are nullable on the wire
    if (s.section_code !== null) {
      expect(typeof s.section_code).toBe("string");
    }
    if (s.instructor_name !== null) {
      expect(typeof s.instructor_name).toBe("string");
    }
    expect(typeof s.member_count).toBe("number");
    expect(s.member_count).toBeGreaterThanOrEqual(0);
    expect(Date.parse(s.created_at)).not.toBeNaN();
  });

  test("term filter returns only sections with exact-match term", async ({
    request,
  }) => {
    if (!courseId) {
      test.skip(true, "No seed course available on staging");
      return;
    }

    // Discover a real term value from the unfiltered list -- exact
    // match means we can't hardcode "Spring 2026" and trust seed
    // data ordering.
    const allResp = await request.get(`/api/courses/${courseId}/sections`);
    expect(allResp.ok()).toBeTruthy();
    const all = await allResp.json();
    if (all.sections.length === 0) {
      test.skip(true, "Seed course has no sections to filter");
      return;
    }
    const targetTerm: string = all.sections[0].term;

    const filtered = await request.get(`/api/courses/${courseId}/sections`, {
      params: { term: targetTerm },
    });
    expect(filtered.ok()).toBeTruthy();
    const body = await filtered.json();

    // Every returned row must be exactly the target term. Some
    // rows from `all` may have a different term and so must NOT
    // appear here; rows with the target term MUST appear.
    expect(body.sections.length).toBeGreaterThan(0);
    for (const s of body.sections) {
      expect(s.term).toBe(targetTerm);
    }
  });

  test("term filter with no match returns empty array (not 404)", async ({
    request,
  }) => {
    if (!courseId) {
      test.skip(true, "No seed course available on staging");
      return;
    }

    // Use a deliberately-implausible term value that no seed
    // section is going to have.
    const resp = await request.get(`/api/courses/${courseId}/sections`, {
      params: { term: "Summer 2099" },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.sections).toEqual([]);
  });

  test("sections are sorted by term DESC", async ({ request }) => {
    if (!courseId) {
      test.skip(true, "No seed course available on staging");
      return;
    }

    const resp = await request.get(`/api/courses/${courseId}/sections`);
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();

    // Need at least 2 sections with distinct terms to validate
    // ordering. Test data may not have that; skip if so.
    const distinctTerms = Array.from(
      new Set(body.sections.map((s: { term: string }) => s.term)),
    ) as string[];
    if (distinctTerms.length < 2) {
      test.skip(
        true,
        `Seed course has fewer than 2 distinct terms (${distinctTerms.length}); cannot validate term DESC ordering`,
      );
      return;
    }

    // Walk pairs of consecutive rows: the term of an earlier row
    // must lex-compare >= the term of any later row (DESC). This
    // is the same lex-comparison the SQL ORDER BY uses, so the
    // assertion exactly mirrors server-side behavior.
    for (let i = 1; i < body.sections.length; i++) {
      const prev: string = body.sections[i - 1].term;
      const cur: string = body.sections[i].term;
      expect(prev >= cur).toBeTruthy();
    }
  });

  test("empty string term is rejected at the wrapper", async ({ request }) => {
    if (!courseId) {
      test.skip(true, "No seed course available on staging");
      return;
    }

    // The kin-openapi wrapper enforces `allowEmptyValue: false`
    // (the default for query params) and rejects `?term=` with
    // 400 BEFORE the handler runs. The service-side
    // empty-string-collapses-to-no-filter logic is defense-in-depth
    // for internal Go callers that bypass the wrapper -- not
    // reachable via the HTTP boundary.
    //
    // This test pins the wire behavior so a future "fix" that
    // drops the empty-string guard from the wrapper would be
    // caught here.
    const resp = await request.get(`/api/courses/${courseId}/sections`, {
      params: { term: "" },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body).toHaveProperty("details");
    expect(body.details).toHaveProperty("term");
  });
});
