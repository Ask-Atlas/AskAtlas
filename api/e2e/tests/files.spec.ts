import { test, expect } from "@playwright/test";

test.describe("Files API", () => {
  const VALID_SORTS = [
    "updated_at",
    "created_at",
    "name",
    "size",
    "status",
    "mime_type",
  ];
  const VALID_SCOPES = ["owned"];
  const VALID_STATUSES = ["pending", "complete", "failed"];

  test("GET /files returns correct structure and data types", async ({
    request,
  }) => {
    const response = await request.get("/api/me/files", {
      params: { scope: "owned", page_limit: 5 },
    });
    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(body).toHaveProperty("files");
    expect(body).toHaveProperty("has_more");
    expect(body).toHaveProperty("next_cursor");
    expect(Array.isArray(body.files)).toBeTruthy();

    if (body.files.length > 0) {
      const file = body.files[0];
      expect(typeof file.id).toBe("string");
      expect(typeof file.name).toBe("string");
      expect(typeof file.size).toBe("number");
      expect(typeof file.mime_type).toBe("string");
      expect(typeof file.status).toBe("string");
      expect(Date.parse(file.created_at)).not.toBeNaN();
      expect(Date.parse(file.updated_at)).not.toBeNaN();
    }
  });

  for (const scope of VALID_SCOPES) {
    test(`GET /files accepts valid scope: ${scope}`, async ({ request }) => {
      const response = await request.get("/api/me/files", {
        params: { scope, page_limit: 1 },
      });
      expect(response.ok()).toBeTruthy();
    });
  }

  test("GET /files rejects invalid scope", async ({ request }) => {
    const response = await request.get("/api/me/files", {
      params: { scope: "invalid_scope" },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body).toHaveProperty("code", 400);
    expect(body).toHaveProperty("status", "Bad Request");
    expect(body.details).toHaveProperty("scope");
    expect(body.details.scope).toContain(
      "must be one of: owned, course, study_guide, accessible",
    );
  });

  for (const sortBy of VALID_SORTS) {
    test(`GET /files accepts valid sort_by: ${sortBy}`, async ({ request }) => {
      const response = await request.get("/api/me/files", {
        params: { scope: "owned", sort_by: sortBy, page_limit: 1 },
      });
      expect(response.ok()).toBeTruthy();
    });
  }

  test("GET /files rejects invalid sort_by", async ({ request }) => {
    const response = await request.get("/api/me/files", {
      params: { sort_by: "coolness" },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details).toHaveProperty("sort_by");
    expect(body.details.sort_by).toContain("must be one of: updated_at");
  });

  test("GET /files accepts valid sort_dir (asc/desc)", async ({ request }) => {
    const r1 = await request.get("/api/me/files", {
      params: { sort_dir: "asc", page_limit: 1 },
    });
    expect(r1.ok()).toBeTruthy();

    const r2 = await request.get("/api/me/files", {
      params: { sort_dir: "desc", page_limit: 1 },
    });
    expect(r2.ok()).toBeTruthy();
  });

  test("GET /files rejects invalid sort_dir", async ({ request }) => {
    const response = await request.get("/api/me/files", {
      params: { sort_dir: "sideways" },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details).toHaveProperty("sort_dir");
    expect(body.details.sort_dir).toContain("must be one of: asc, desc");
  });

  for (const status of VALID_STATUSES) {
    test(`GET /files filters by valid status: ${status}`, async ({
      request,
    }) => {
      const response = await request.get("/api/me/files", {
        params: { status, page_limit: 1 },
      });
      expect(response.ok()).toBeTruthy();
    });
  }

  test("GET /files rejects invalid status", async ({ request }) => {
    const response = await request.get("/api/me/files", {
      params: { status: "kinda_done" },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details).toHaveProperty("status");
    expect(body.details.status).toContain(
      "must be one of: pending, complete, failed",
    );
  });

  test("GET /files rejects invalid mime_type", async ({ request }) => {
    const response = await request.get("/api/me/files", {
      params: { mime_type: "video/mp4" },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details).toHaveProperty("mime_type");
  });

  test("GET /files validates min_size and max_size", async ({ request }) => {
    const r1 = await request.get("/api/me/files", {
      params: { min_size: 100, max_size: 200, page_limit: 1 },
    });
    expect(r1.ok()).toBeTruthy();

    const r2 = await request.get("/api/me/files", {
      params: { min_size: 200, max_size: 100 },
    });
    expect(r2.status()).toBe(400);
    const b2 = await r2.json();
    expect(b2.details).toHaveProperty("min_size");
    expect(b2.details.min_size).toContain(
      "min_size cannot be greater than max_size",
    );

    const r3 = await request.get("/api/me/files", {
      params: { min_size: -1 },
    });
    expect(r3.status()).toBe(400);
  });

  test("GET /files validates date range logic (created_from/to)", async ({
    request,
  }) => {
    const now = new Date().toISOString();
    const yesterday = new Date(Date.now() - 86400000).toISOString();

    const r1 = await request.get("/api/me/files", {
      params: { created_from: now, created_to: yesterday },
    });
    expect(r1.status()).toBe(400);
    const b1 = await r1.json();
    expect(b1.details).toHaveProperty("created_from");
    expect(b1.details.created_from).toContain("cannot be after created_to");
  });

  test("Pagination: cursor-based traversal returns unique pages", async ({
    request,
  }) => {
    const resp1 = await request.get("/api/me/files", {
      params: { scope: "owned", page_limit: 1 },
    });
    expect(resp1.ok()).toBeTruthy();
    const body1 = await resp1.json();

    if (body1.files.length === 0) {
      test.skip(true, "No files available to test pagination traversal");
      return;
    }

    expect(body1.files).toHaveLength(1);
    expect(body1.has_more).toBeTruthy();
    expect(body1.next_cursor).toBeTruthy();

    const resp2 = await request.get("/api/me/files", {
      params: { scope: "owned", page_limit: 1, cursor: body1.next_cursor },
    });
    expect(resp2.ok()).toBeTruthy();
    const body2 = await resp2.json();

    expect(body2.files).toHaveLength(1);
    expect(body1.files[0].id).not.toBe(body2.files[0].id);
  });

  test("Pagination: invalid page_limit returns error", async ({ request }) => {
    const r1 = await request.get("/api/me/files", {
      params: { page_limit: 1001 },
    });
    expect(r1.status()).toBe(400);
    const b1 = await r1.json();
    expect(b1.details).toHaveProperty("page_limit");

    const r2 = await request.get("/api/me/files", {
      params: { page_limit: -5 },
    });
    expect(r2.status()).toBe(400);
  });

  test("Pagination: invalid cursor returns error", async ({ request }) => {
    const response = await request.get("/api/me/files", {
      params: { cursor: "invalid-base64-cursor" },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details).toHaveProperty("cursor");
  });

  test.describe("GET /files/:file_id", () => {
    test("returns an existing file", async ({ request }) => {
      const listResponse = await request.get("/api/me/files", {
        params: { scope: "owned", page_limit: 1 },
      });
      expect(listResponse.ok()).toBeTruthy();
      const listBody = await listResponse.json();

      if (listBody.files.length === 0) {
        test.skip(true, "No files available to test GetFile");
        return;
      }

      const existingFile = listBody.files[0];
      const response = await request.get(`/api/files/${existingFile.id}`);
      expect(response.ok()).toBeTruthy();
      const file = await response.json();

      expect(file.id).toBe(existingFile.id);
      expect(file.name).toBe(existingFile.name);
      expect(file.size).toBe(existingFile.size);
      expect(file.mime_type).toBe(existingFile.mime_type);
      expect(file.status).toBe(existingFile.status);
    });

    test("returns 404 for non-existent ID", async ({ request }) => {
      const response = await request.get(
        "/api/files/00000000-0000-0000-0000-000000000000",
      );
      expect(response.status()).toBe(404);
    });
  });
});
