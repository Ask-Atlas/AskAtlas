import { test, expect } from "@playwright/test";

// E2E coverage for study-guide grant endpoints (ASK-211):
//
//   GET    /api/study-guides/{id}/grants  -- list grants (creator-only)
//   POST   /api/study-guides/{id}/grants  -- create grant (creator-only)
//   DELETE /api/study-guides/{id}/grants  -- revoke grant (creator-only)
//
// Two suites:
//
// 1. Non-destructive wire-contract checks (401 / 400 / 404) against
//    NIL_UUID. Mirrors files-grants.spec.ts -- verifies chi +
//    oapi-codegen + service enum validation before any DB write.
//
// 2. Single-token lifecycle: create a private study guide, list its
//    auto-seeded course grant, PATCH visibility, revoke the course
//    grant, verify revoke-again returns 404, soft-delete the guide.
//    Exercises the full CRUD loop against real staging DB rows. A
//    cross-user 404-for-non-creator assertion requires a second
//    token and is left for the two-token harness.

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

const validBody = {
  grantee_type: "user",
  grantee_id: NIL_UUID,
  permission: "view",
};

test.describe("CreateStudyGuideGrant validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.post(`/api/study-guides/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid study-guide UUID with 400", async ({ request }) => {
    const resp = await request.post("/api/study-guides/not-a-uuid/grants", {
      data: validBody,
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid grantee_type with 400", async ({ request }) => {
    const resp = await request.post(`/api/study-guides/${NIL_UUID}/grants`, {
      data: { ...validBody, grantee_type: "organization" },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects grantee_type=study_guide with 400", async ({ request }) => {
    // The study_guide_grants CHECK pins grantee_type to ('user','course');
    // the service must also refuse 'study_guide' before the DB fires.
    const resp = await request.post(`/api/study-guides/${NIL_UUID}/grants`, {
      data: { ...validBody, grantee_type: "study_guide" },
    });
    expect(resp.status()).toBe(400);
  });

  test("rejects invalid permission with 400", async ({ request }) => {
    const resp = await request.post(`/api/study-guides/${NIL_UUID}/grants`, {
      data: { ...validBody, permission: "whatever" },
    });
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent study guide", async ({ request }) => {
    const resp = await request.post(`/api/study-guides/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(404);
  });
});

test.describe("RevokeStudyGuideGrant validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.delete(`/api/study-guides/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid study-guide UUID with 400", async ({ request }) => {
    const resp = await request.delete("/api/study-guides/not-a-uuid/grants", {
      data: validBody,
    });
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent study guide", async ({ request }) => {
    const resp = await request.delete(`/api/study-guides/${NIL_UUID}/grants`, {
      data: validBody,
    });
    expect(resp.status()).toBe(404);
  });
});

test.describe("ListStudyGuideGrants validation", () => {
  test("rejects unauthenticated with 401", async ({ playwright }) => {
    const noAuth = await playwright.request.newContext({
      baseURL: process.env.E2E_BASE_URL,
      extraHTTPHeaders: {},
    });
    const resp = await noAuth.get(`/api/study-guides/${NIL_UUID}/grants`);
    expect(resp.status()).toBe(401);
    await noAuth.dispose();
  });

  test("rejects invalid study-guide UUID with 400", async ({ request }) => {
    const resp = await request.get("/api/study-guides/not-a-uuid/grants");
    expect(resp.status()).toBe(400);
  });

  test("returns 404 for non-existent study guide", async ({ request }) => {
    const resp = await request.get(`/api/study-guides/${NIL_UUID}/grants`);
    expect(resp.status()).toBe(404);
  });
});

test.describe("Study-guide grants lifecycle (single-token)", () => {
  // Requires E2E_COURSE_ID to point at a course the token's user is
  // enrolled in (course preflight otherwise 404s on create). If
  // absent, skip the whole suite.
  const courseID = process.env.E2E_COURSE_ID;

  test.skip(
    !courseID,
    "E2E_COURSE_ID not set; skipping destructive lifecycle",
  );

  test("create -> list seeded grant -> duplicate 409 -> revoke -> 404 on revoke-again -> patch visibility -> cleanup", async ({
    request,
  }) => {
    // 1. Create a private study guide. The service auto-seeds a
    //    course grant so course members can still see it.
    const title = `ASK-211 e2e ${new Date().toISOString()}`;
    const createResp = await request.post(
      `/api/courses/${courseID}/study-guides`,
      {
        data: {
          title,
          description: "temporary fixture for grants lifecycle test",
          visibility: "private",
        },
      },
    );
    expect(createResp.status()).toBe(201);
    const created = await createResp.json();
    const sgID = created.id as string;
    expect(created).toHaveProperty("visibility", "private");

    try {
      // 2. GET /grants -- the auto-seeded course grant must appear.
      const listResp = await request.get(`/api/study-guides/${sgID}/grants`);
      expect(listResp.status()).toBe(200);
      const listed = await listResp.json();
      expect(Array.isArray(listed.grants)).toBe(true);
      const courseGrant = listed.grants.find(
        (g: { grantee_type: string; grantee_id: string }) =>
          g.grantee_type === "course" && g.grantee_id === courseID,
      );
      expect(
        courseGrant,
        "course grant auto-seeded on private guide create",
      ).toBeTruthy();

      // 3. Duplicate the auto-seeded course grant -> 409 Conflict.
      const dupResp = await request.post(
        `/api/study-guides/${sgID}/grants`,
        {
          data: {
            grantee_type: "course",
            grantee_id: courseID,
            permission: "view",
          },
        },
      );
      expect(dupResp.status()).toBe(409);

      // 4. DELETE the auto-seeded course grant -> 204.
      const revokeResp = await request.delete(
        `/api/study-guides/${sgID}/grants`,
        {
          data: {
            grantee_type: "course",
            grantee_id: courseID,
            permission: "view",
          },
        },
      );
      expect(revokeResp.status()).toBe(204);

      // 5. DELETE again -> 404 (not idempotent, spec parity with
      //    file_grants).
      const revokeAgainResp = await request.delete(
        `/api/study-guides/${sgID}/grants`,
        {
          data: {
            grantee_type: "course",
            grantee_id: courseID,
            permission: "view",
          },
        },
      );
      expect(revokeAgainResp.status()).toBe(404);

      // 6. PATCH visibility=public -> 200 + the field round-trips.
      const patchResp = await request.patch(`/api/study-guides/${sgID}`, {
        data: { visibility: "public" },
      });
      expect(patchResp.status()).toBe(200);
      const patched = await patchResp.json();
      expect(patched).toHaveProperty("visibility", "public");
    } finally {
      // 7. Cleanup -- soft-delete the fixture so we don't leave
      //    stray rows on staging.
      await request.delete(`/api/study-guides/${sgID}`);
    }
  });
});
