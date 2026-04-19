import { test, expect } from "@playwright/test";

// E2E coverage for the practice-sessions surface.
//
// Currently only ASK-149 (GET /quizzes/{quiz_id}/sessions) is
// covered; sibling endpoints (POST start, POST submit, POST
// complete, GET by id) get added here as those tickets ship and
// need staging verification beyond the Go unit/handler tests.
//
// Discovery: rather than pinning a hardcoded quiz id (which
// staging seed data doesn't currently expose), the substantive
// happy-path tests walk
// courses -> study-guides -> quizzes
// once in beforeAll, then start a session on the first quiz
// found. Tests skip cleanly when staging has no seeded
// course/guide/quiz, mirroring the files.spec.ts convention.
//
// The validation tests do NOT need real seed data -- they only
// hit error paths (auth, malformed UUID, range violations,
// invalid cursor), so they always run.

test.describe("Sessions API (ASK-149 list)", () => {
  // Shared discovery state across all happy-path tests in this
  // describe. Set in beforeAll; null when no seed data was found
  // (each substantive test then early-skips).
  let quizId: string | null = null;
  let startedSessionId: string | null = null;

  test.beforeAll(async ({ request }) => {
    // 1. Find a course
    const coursesResp = await request.get("/api/courses", {
      params: { page_limit: 1 },
    });
    if (!coursesResp.ok()) return;
    const courses = await coursesResp.json();
    const course = courses?.courses?.[0];
    if (!course?.id) return;

    // 2. Find a study guide on that course
    const guidesResp = await request.get(
      `/api/courses/${course.id}/study-guides`,
      { params: { page_limit: 1 } },
    );
    if (!guidesResp.ok()) return;
    const guides = await guidesResp.json();
    const guide = guides?.study_guides?.[0];
    if (!guide?.id) return;

    // 3. Find a quiz on that study guide
    const quizzesResp = await request.get(
      `/api/study-guides/${guide.id}/quizzes`,
    );
    if (!quizzesResp.ok()) return;
    const quizzes = await quizzesResp.json();
    const quiz = quizzes?.quizzes?.[0];
    if (!quiz?.id) return;
    quizId = quiz.id;

    // 4. Ensure the test user has at least one session on this
    //    quiz (idempotent: returns 200 with the existing
    //    in-progress session if one is already there, 201 if
    //    fresh). Either way, the listing must surface it.
    const startResp = await request.post(
      `/api/quizzes/${quizId}/sessions`,
    );
    if (!startResp.ok()) return;
    const session = await startResp.json();
    startedSessionId = session?.id ?? null;
  });

  // ---------- Validation (no seed data required) ----------

  test("rejects unauthenticated", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get(
      "/api/quizzes/00000000-0000-0000-0000-000000000000/sessions",
    );
    expect([401, 403]).toContain(resp.status());
    await noAuth.dispose();
  });

  test("rejects invalid quiz UUID with 400", async ({ request }) => {
    const resp = await request.get("/api/quizzes/not-a-uuid/sessions");
    expect(resp.status()).toBe(400);
  });

  test("rejects limit=0 with 400", async ({ request }) => {
    const resp = await request.get(
      "/api/quizzes/00000000-0000-0000-0000-000000000000/sessions",
      { params: { limit: 0 } },
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects limit=51 with 400", async ({ request }) => {
    const resp = await request.get(
      "/api/quizzes/00000000-0000-0000-0000-000000000000/sessions",
      { params: { limit: 51 } },
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects unknown status with 400", async ({ request }) => {
    const resp = await request.get(
      "/api/quizzes/00000000-0000-0000-0000-000000000000/sessions",
      { params: { status: "pending" } },
    );
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.details?.status).toContain("is not one of the allowed values");
  });

  test("rejects garbled cursor with details.cursor", async ({ request }) => {
    // The 400 may come from the wrapper layer (when the cursor's
    // own format is rejected before our handler runs). When the
    // wrapper accepts the value but the handler can't decode it,
    // we get our typed { details: { cursor: "..." } }. Accept
    // either path -- both are valid 400s for the spec's purpose.
    const resp = await request.get(
      "/api/quizzes/00000000-0000-0000-0000-000000000000/sessions",
      { params: { cursor: "definitely-not-base64-$$$" } },
    );
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for a non-existent quiz", async ({ request }) => {
    const resp = await request.get(
      "/api/quizzes/00000000-0000-0000-0000-000000000000/sessions",
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body).toHaveProperty("status", "Not Found");
  });

  // ---------- Happy path (requires seed data) ----------

  test("returns 200 + ListSessionsResponse shape on default params", async ({
    request,
  }) => {
    if (!quizId) {
      test.skip(true, "No seed quiz available on staging");
      return;
    }

    const resp = await request.get(`/api/quizzes/${quizId}/sessions`);
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();

    expect(body).toHaveProperty("sessions");
    expect(Array.isArray(body.sessions)).toBeTruthy();
    expect(typeof body.has_more).toBe("boolean");
    // next_cursor follows the same logical-tie as ListFilesResponse:
    // when has_more is true, the cursor MUST be a string; when
    // false, it must be null or absent.
    if (body.has_more) {
      expect(typeof body.next_cursor).toBe("string");
    } else {
      expect(body.next_cursor ?? null).toBeNull();
    }

    // The session we started in beforeAll must appear in the
    // listing -- the list is the user's own scope.
    if (startedSessionId) {
      const ids = body.sessions.map((s: { id: string }) => s.id);
      expect(ids).toContain(startedSessionId);
    }

    // Per-row shape on at least the first session.
    if (body.sessions.length > 0) {
      const s = body.sessions[0];
      expect(s.id).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
      );
      expect(Date.parse(s.started_at)).not.toBeNaN();
      expect(typeof s.total_questions).toBe("number");
      expect(typeof s.correct_answers).toBe("number");
      // completed_at + score_percentage gate together: both null
      // for in-progress, both set for completed.
      if (s.completed_at === null) {
        expect(s.score_percentage).toBeNull();
      } else {
        expect(Date.parse(s.completed_at)).not.toBeNaN();
        expect(typeof s.score_percentage).toBe("number");
        expect(s.score_percentage).toBeGreaterThanOrEqual(0);
        expect(s.score_percentage).toBeLessThanOrEqual(100);
      }
    }
  });

  test("status=active returns only in-progress sessions", async ({
    request,
  }) => {
    if (!quizId) {
      test.skip(true, "No seed quiz available on staging");
      return;
    }

    const resp = await request.get(`/api/quizzes/${quizId}/sessions`, {
      params: { status: "active" },
    });
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();

    for (const s of body.sessions) {
      expect(s.completed_at).toBeNull();
      expect(s.score_percentage).toBeNull();
    }

    // beforeAll just started a session, so this must include it.
    if (startedSessionId) {
      const ids = body.sessions.map((s: { id: string }) => s.id);
      expect(ids).toContain(startedSessionId);
    }
  });

  test("status=completed returns only finalised sessions", async ({
    request,
  }) => {
    if (!quizId) {
      test.skip(true, "No seed quiz available on staging");
      return;
    }

    const resp = await request.get(`/api/quizzes/${quizId}/sessions`, {
      params: { status: "completed" },
    });
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();

    for (const s of body.sessions) {
      expect(s.completed_at).not.toBeNull();
      expect(Date.parse(s.completed_at)).not.toBeNaN();
      expect(typeof s.score_percentage).toBe("number");
      expect(s.score_percentage).toBeGreaterThanOrEqual(0);
      expect(s.score_percentage).toBeLessThanOrEqual(100);
    }
  });

  test("limit=1 + cursor follow returns distinct pages", async ({
    request,
  }) => {
    if (!quizId) {
      test.skip(true, "No seed quiz available on staging");
      return;
    }

    const first = await request.get(`/api/quizzes/${quizId}/sessions`, {
      params: { limit: 1 },
    });
    expect(first.ok()).toBeTruthy();
    const firstBody = await first.json();
    expect(firstBody.sessions.length).toBeLessThanOrEqual(1);

    if (!firstBody.has_more) {
      test.skip(
        true,
        "Test user has fewer than 2 sessions on this quiz; cannot test cursor follow",
      );
      return;
    }

    expect(typeof firstBody.next_cursor).toBe("string");

    const second = await request.get(`/api/quizzes/${quizId}/sessions`, {
      params: { limit: 1, cursor: firstBody.next_cursor },
    });
    expect(second.ok()).toBeTruthy();
    const secondBody = await second.json();

    expect(secondBody.sessions).toHaveLength(1);
    expect(firstBody.sessions[0].id).not.toBe(secondBody.sessions[0].id);
  });

  test("sessions are sorted by started_at DESC", async ({ request }) => {
    if (!quizId) {
      test.skip(true, "No seed quiz available on staging");
      return;
    }

    const resp = await request.get(`/api/quizzes/${quizId}/sessions`, {
      params: { limit: 50 },
    });
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();

    if (body.sessions.length < 2) {
      test.skip(
        true,
        "Need at least 2 sessions on this quiz to validate ordering",
      );
      return;
    }

    for (let i = 1; i < body.sessions.length; i++) {
      const prev = Date.parse(body.sessions[i - 1].started_at);
      const cur = Date.parse(body.sessions[i].started_at);
      expect(prev).toBeGreaterThanOrEqual(cur);
    }
  });
});
