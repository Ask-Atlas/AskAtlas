import { test, expect } from "@playwright/test";

// E2E coverage for quiz-question endpoints (ASK-119 + ASK-108):
//
//   DELETE /api/quizzes/{quiz_id}/questions/{question_id} -- ASK-119
//   PUT    /api/quizzes/{quiz_id}/questions/{question_id} -- ASK-108
//
// Service-layer tests in api/internal/quizzes/service_test.go cover
// the full matrix (success / 403 not-creator / 404 missing / 404
// sibling-quiz / 400 last-question / 400 validation / type-change
// replacements) against mocked repo expectations. This file
// backfills the missing wire-contract verification against dev/
// staging for the 401/400/404 paths -- the surface chi +
// oapi-codegen + service enforce before any DB write fires.
//
// **Strictly non-destructive coverage**: every test targets the NIL
// UUID for both quiz_id and question_id, so no real question is
// ever deleted or replaced by this suite. The 200/204 happy paths
// and the 400 last-question guard require real rows, so they live
// in the Go service tests.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

// validReplaceBody mirrors the minimum valid PUT body from ASK-108 --
// a well-formed MCQ with 4 options and exactly one correct. The
// tests below never submit this against a real quiz because quiz_id
// is NIL_UUID, but the body must be valid so the service reaches
// the 404 path rather than rejecting at validation.
const validReplaceBody = {
  type: "multiple-choice",
  question: "Which traversal visits the root node first?",
  options: [
    { text: "In-order", is_correct: false },
    { text: "Pre-order", is_correct: true },
    { text: "Post-order", is_correct: false },
    { text: "Level-order", is_correct: false },
  ],
};

test.describe("DeleteQuizQuestion validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(
      `/api/quizzes/${NIL_UUID}/questions/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid quiz_id with 400", async ({ request }) => {
    const resp = await request.delete(
      `/api/quizzes/not-a-uuid/questions/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid question_id with 400", async ({ request }) => {
    const resp = await request.delete(
      `/api/quizzes/${NIL_UUID}/questions/not-a-uuid`,
    );
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent quiz", async ({ request }) => {
    // NIL_UUID never matches a real quiz, so the service's locked
    // SELECT returns sql.ErrNoRows -> 404 before any DELETE runs.
    const resp = await request.delete(
      `/api/quizzes/${NIL_UUID}/questions/${NIL_UUID}`,
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});

test.describe("ReplaceQuizQuestion validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.put(
      `/api/quizzes/${NIL_UUID}/questions/${NIL_UUID}`,
      { data: validReplaceBody },
    );
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid quiz_id with 400", async ({ request }) => {
    const resp = await request.put(
      `/api/quizzes/not-a-uuid/questions/${NIL_UUID}`,
      { data: validReplaceBody },
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid question_id with 400", async ({ request }) => {
    const resp = await request.put(
      `/api/quizzes/${NIL_UUID}/questions/not-a-uuid`,
      { data: validReplaceBody },
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects missing type field with 400", async ({ request }) => {
    const { type: _omit, ...bodyMissingType } = validReplaceBody;
    void _omit;
    const resp = await request.put(
      `/api/quizzes/${NIL_UUID}/questions/${NIL_UUID}`,
      { data: bodyMissingType },
    );
    expect(resp.status()).toBe(400);
  });

  test("rejects MCQ with zero correct options with 400", async ({
    request,
  }) => {
    const resp = await request.put(
      `/api/quizzes/${NIL_UUID}/questions/${NIL_UUID}`,
      {
        data: {
          ...validReplaceBody,
          options: validReplaceBody.options.map((o) => ({
            ...o,
            is_correct: false,
          })),
        },
      },
    );
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.details?.options).toMatch(/exactly one/);
  });

  test("returns 404 for non-existent quiz", async ({ request }) => {
    // Valid body, well-formed UUID format, but NIL_UUID resolves
    // to nothing -- service's locked SELECT -> 404 before any
    // UPDATE/INSERT runs.
    const resp = await request.put(
      `/api/quizzes/${NIL_UUID}/questions/${NIL_UUID}`,
      { data: validReplaceBody },
    );
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});
