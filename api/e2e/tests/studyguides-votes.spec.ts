import { test, expect } from "@playwright/test";

// E2E coverage for the study-guide voting pair:
//   - POST   /api/study-guides/{study_guide_id}/votes  (ASK-139)
//   - DELETE /api/study-guides/{study_guide_id}/votes  (ASK-141)
//
// Both endpoints landed together in PR #139 with Go unit + handler
// tests; this file backfills the missing wire-contract verification
// against dev/staging so refactors that drift either response shape
// trip a deterministic e2e failure rather than silently shipping.
//
// Discovery: validation tests run unconditionally; happy-path tests
// look up any visible study guide via
// GET /courses/{course_id}/study-guides and skip cleanly when no
// seed data exists.
//
// Serialization: substantive tests share per-(user, guide) vote
// state guarded by the (user_id, study_guide_id) primary key on
// study_guide_votes. Two parallel tests would race for the same
// slot, so the lifecycle describe runs in serial mode and resets
// state with a best-effort DELETE before each test.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

// Vote score envelopes that may be either int or string-encoded
// depending on how the test runtime serialized the JSON response.
// Empirically pgx returns a Go int64 which the api.gen.go marshals
// as a JSON number; this normalizer keeps the assertions robust to
// future int64-string codec swaps without weakening the contract.
function asNumber(v: unknown): number {
  if (typeof v === "number") return v;
  if (typeof v === "string") return Number(v);
  throw new Error(`expected vote_score number, got ${typeof v}: ${String(v)}`);
}

test.describe("StudyGuideVotes validation", () => {
  test("POST rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.post(
      `/api/study-guides/${NIL_UUID}/votes`,
      { data: { vote: "up" } },
    );
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("DELETE rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(`/api/study-guides/${NIL_UUID}/votes`);
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("POST rejects invalid guide UUID with 400", async ({ request }) => {
    const resp = await request.post("/api/study-guides/not-a-uuid/votes", {
      data: { vote: "up" },
    });
    expect(resp.status()).toBe(400);
  });

  test("DELETE rejects invalid guide UUID with 400", async ({ request }) => {
    const resp = await request.delete("/api/study-guides/not-a-uuid/votes");
    expect(resp.status()).toBe(400);
  });

  test("POST rejects invalid vote enum with 400", async ({ request }) => {
    const resp = await request.post(`/api/study-guides/${NIL_UUID}/votes`, {
      data: { vote: "sideways" },
    });
    expect(resp.status()).toBe(400);
  });

  test("POST returns 404 for a non-existent guide", async ({ request }) => {
    const resp = await request.post(`/api/study-guides/${NIL_UUID}/votes`, {
      data: { vote: "up" },
    });
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });

  test("DELETE returns 404 for a non-existent guide", async ({ request }) => {
    const resp = await request.delete(`/api/study-guides/${NIL_UUID}/votes`);
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });
});

// Lifecycle tests share per-(viewer, guide) vote state via the PK
// on study_guide_votes; serial mode keeps them deterministic.
test.describe("StudyGuideVotes lifecycle", () => {
  test.describe.configure({ mode: "serial" });

  // Shared discovery state. Set in beforeAll; null when no seed
  // data was found (substantive tests early-skip).
  let guideId: string | null = null;

  test.beforeAll(async ({ request }) => {
    const coursesResp = await request.get("/api/courses", {
      params: { page_limit: 1 },
    });
    if (!coursesResp.ok()) return;
    const course = (await coursesResp.json())?.courses?.[0];
    if (!course?.id) return;

    const guidesResp = await request.get(
      `/api/courses/${course.id}/study-guides`,
      { params: { page_limit: 1 } },
    );
    if (!guidesResp.ok()) return;
    const guide = (await guidesResp.json())?.study_guides?.[0];
    if (!guide?.id) return;
    guideId = guide.id;
  });

  // Reset per-(viewer, guide) vote state before each lifecycle
  // assertion so the suite is order-independent and a previous
  // test's leftover vote can't leak into the next one. 204 means
  // we cleared a stale row; 404 means there was nothing to clear
  // -- both are valid starting points for the next test.
  test.beforeEach(async ({ request }) => {
    if (!guideId) return;
    const resp = await request.delete(`/api/study-guides/${guideId}/votes`);
    expect([204, 404]).toContain(resp.status());
  });

  test("POST upvote returns 200 with CastVoteResponse shape", async ({
    request,
  }) => {
    if (!guideId) {
      test.skip(true, "No seed study guide available on staging");
      return;
    }

    const resp = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "up" },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    expect(body).toHaveProperty("vote", "up");
    expect(body).toHaveProperty("vote_score");
    // vote_score is always an integer; sign depends on existing
    // votes by other users on this guide, so we only assert the
    // wire type, not a value.
    const score = asNumber(body.vote_score);
    expect(Number.isInteger(score)).toBeTruthy();
  });

  test("POST same direction is idempotent (no score drift)", async ({
    request,
  }) => {
    if (!guideId) {
      test.skip(true, "No seed study guide available on staging");
      return;
    }

    const first = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "up" },
    });
    expect(first.status()).toBe(200);
    const firstScore = asNumber((await first.json()).vote_score);

    const second = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "up" },
    });
    expect(second.status()).toBe(200);
    const secondBody = await second.json();
    expect(secondBody.vote).toBe("up");
    expect(asNumber(secondBody.vote_score)).toBe(firstScore);
  });

  test("POST opposite direction flips vote and shifts score by 2", async ({
    request,
  }) => {
    if (!guideId) {
      test.skip(true, "No seed study guide available on staging");
      return;
    }

    const up = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "up" },
    });
    expect(up.status()).toBe(200);
    const upScore = asNumber((await up.json()).vote_score);

    const down = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "down" },
    });
    expect(down.status()).toBe(200);
    const downBody = await down.json();
    expect(downBody.vote).toBe("down");
    // up -> down: viewer's contribution goes from +1 to -1, so the
    // aggregate must drop by exactly 2 regardless of other users'
    // votes (which stay constant during the test).
    expect(asNumber(downBody.vote_score)).toBe(upScore - 2);
  });

  test("DELETE after POST returns 204 and re-cast restores score", async ({
    request,
  }) => {
    if (!guideId) {
      test.skip(true, "No seed study guide available on staging");
      return;
    }

    const cast = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "up" },
    });
    expect(cast.status()).toBe(200);
    const castScore = asNumber((await cast.json()).vote_score);

    const remove = await request.delete(
      `/api/study-guides/${guideId}/votes`,
    );
    expect(remove.status()).toBe(204);
    expect(await remove.text()).toBe("");

    // Re-cast to read back the new aggregate. This is the only
    // GET-style probe of vote_score available via the votes API
    // surface; the alternative (GET /study-guides/{id}) carries
    // the same field but pulls a far heavier payload.
    const recast = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "up" },
    });
    expect(recast.status()).toBe(200);
    const recastScore = asNumber((await recast.json()).vote_score);
    // After delete the aggregate dropped by 1 (viewer's +1 removed).
    // After re-up the aggregate gained 1 back. Net: equal to castScore.
    expect(recastScore).toBe(castScore);
  });

  test("DELETE without an existing vote returns 404", async ({ request }) => {
    if (!guideId) {
      test.skip(true, "No seed study guide available on staging");
      return;
    }

    // beforeEach already cleared any prior vote, so calling DELETE
    // here without a preceding POST exercises the "no existing
    // vote" branch.
    const resp = await request.delete(`/api/study-guides/${guideId}/votes`);
    expect(resp.status()).toBe(404);
    const body = await resp.json();
    expect(body).toHaveProperty("code", 404);
    expect(body.message).toMatch(/not found/i);
  });

  test("double DELETE: first 204, second 404", async ({ request }) => {
    if (!guideId) {
      test.skip(true, "No seed study guide available on staging");
      return;
    }

    // Establish a vote.
    const cast = await request.post(`/api/study-guides/${guideId}/votes`, {
      data: { vote: "down" },
    });
    expect(cast.status()).toBe(200);

    const first = await request.delete(`/api/study-guides/${guideId}/votes`);
    expect(first.status()).toBe(204);

    const second = await request.delete(`/api/study-guides/${guideId}/votes`);
    expect(second.status()).toBe(404);
    const body = await second.json();
    expect(body.message).toMatch(/not found/i);
  });
});
