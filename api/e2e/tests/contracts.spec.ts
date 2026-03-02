import { test, expect } from "@playwright/test";

test.describe("API Contracts & Error Handling", () => {
  test.describe("Success Contracts", () => {
    test("GET /files matches FileResponse DTO structure", async ({
      request,
    }) => {
      const response = await request.get("/api/me/files", {
        params: { scope: "owned", page_limit: 1 },
      });
      expect(response.ok()).toBeTruthy();
      const body = await response.json();

      expect(body).toHaveProperty("files");
      expect(body).toHaveProperty("has_more");
      expect(typeof body.has_more).toBe("boolean");

      if (body.next_cursor !== undefined && body.next_cursor !== null) {
        expect(typeof body.next_cursor).toBe("string");
      }

      if (body.files.length > 0) {
        const file = body.files[0];

        expect(file).toHaveProperty("id");
        expect(file.id).toMatch(
          /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
        );

        expect(file).toHaveProperty("name");
        expect(typeof file.name).toBe("string");

        expect(file).toHaveProperty("mime_type");
        expect(typeof file.mime_type).toBe("string");

        expect(file).toHaveProperty("status");
        expect(typeof file.status).toBe("string");
        expect(["pending", "complete", "failed"]).toContain(file.status);

        expect(file).toHaveProperty("size");
        expect(typeof file.size).toBe("number");

        expect(file).toHaveProperty("created_at");
        expect(Date.parse(file.created_at)).not.toBeNaN();

        expect(file).toHaveProperty("updated_at");
        expect(Date.parse(file.updated_at)).not.toBeNaN();

        if (file.favorited_at !== undefined && file.favorited_at !== null) {
          expect(Date.parse(file.favorited_at)).not.toBeNaN();
        }

        if (file.last_viewed_at !== undefined && file.last_viewed_at !== null) {
          expect(Date.parse(file.last_viewed_at)).not.toBeNaN();
        }
      }
    });
  });

  test.describe("Error Contracts", () => {
    test("Unauthenticated request is rejected", async ({ playwright }) => {
      const apiContext = await playwright.request.newContext({
        baseURL: process.env.E2E_BASE_URL,
        extraHTTPHeaders: {},
      });

      const response = await apiContext.get("/api/me/files");
      expect([401, 403]).toContain(response.status());

      await apiContext.dispose();
    });

    test("404 Not Found structure", async ({ request }) => {
      const response = await request.get(
        "/api/files/00000000-0000-0000-0000-000000000000",
      );

      expect(response.status()).toBe(404);
      const body = await response.json();

      expect(body).toHaveProperty("code", 404);
      expect(body).toHaveProperty("status", "Not Found");
      expect(body).toHaveProperty("message");
      expect(typeof body.message).toBe("string");
    });

    test("400 Bad Request structure", async ({ request }) => {
      const response = await request.get("/api/me/files", {
        params: { scope: "INVALID_SCOPE" },
      });

      expect(response.status()).toBe(400);
      const body = await response.json();

      expect(body).toHaveProperty("code", 400);
      expect(body).toHaveProperty("status", "Bad Request");
      expect(body).toHaveProperty("message");
      expect(typeof body.message).toBe("string");
      expect(body).toHaveProperty("details");
      expect(typeof body.details).toBe("object");
      expect(body.details).toHaveProperty("scope");
    });

    test("GET /files/:file_id matches FileResponse DTO structure", async ({
      request,
    }) => {
      const listResponse = await request.get("/api/me/files", {
        params: { scope: "owned", page_limit: 1 },
      });
      expect(listResponse.ok()).toBeTruthy();
      const listBody = await listResponse.json();

      if (listBody.files.length === 0) {
        test.skip(true, "No files available to test GetFile contract");
        return;
      }

      const response = await request.get(`/api/files/${listBody.files[0].id}`);
      expect(response.ok()).toBeTruthy();
      const file = await response.json();

      expect(file.id).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
      );
      expect(typeof file.name).toBe("string");
      expect(typeof file.size).toBe("number");
      expect(typeof file.mime_type).toBe("string");
      expect(["pending", "complete", "failed"]).toContain(file.status);
      expect(Date.parse(file.created_at)).not.toBeNaN();
      expect(Date.parse(file.updated_at)).not.toBeNaN();

      if (file.favorited_at !== undefined && file.favorited_at !== null) {
        expect(Date.parse(file.favorited_at)).not.toBeNaN();
      }
      if (file.last_viewed_at !== undefined && file.last_viewed_at !== null) {
        expect(Date.parse(file.last_viewed_at)).not.toBeNaN();
      }
    });

    test("400 Bad Request structure (invalid file_id format)", async ({
      request,
    }) => {
      const response = await request.get("/api/files/not-a-uuid");

      expect(response.status()).toBe(400);
      const body = await response.json();

      expect(body).toHaveProperty("code", 400);
      expect(body).toHaveProperty("status", "Bad Request");
      expect(body).toHaveProperty("message");
      expect(typeof body.message).toBe("string");
    });
  });
});
