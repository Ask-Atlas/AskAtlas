/**
 * Unit tests for the `unwrap` / `unwrapVoid` helpers + `ApiError`.
 *
 * These are deliberately decoupled from openapi-fetch so the logic
 * stays fast + runs under jsdom without a Response polyfill. Tests
 * that exercise the full server-action stack live in the sibling
 * action files' test suites.
 */
import { ApiError, unwrap, unwrapVoid } from "./errors";
import type { AppError } from "./types";

function fakeResponse(status: number): Response {
  return {
    status,
    ok: status >= 200 && status < 300,
  } as unknown as Response;
}

describe("unwrap", () => {
  it("returns the data on success", () => {
    expect(
      unwrap(
        { data: { files: [], has_more: false }, response: fakeResponse(200) },
        "GET /me/files",
      ),
    ).toEqual({ files: [], has_more: false });
  });

  it("throws ApiError with the parsed AppError body when error is set", () => {
    const err: AppError = {
      code: 404,
      status: "not_found",
      message: "file missing",
    };

    let caught: unknown;
    try {
      unwrap({ error: err, response: fakeResponse(404) }, "GET /files/missing");
    } catch (e) {
      caught = e;
    }

    expect(caught).toBeInstanceOf(ApiError);
    expect((caught as ApiError).status).toBe(404);
    expect((caught as ApiError).body).toEqual(err);
    expect((caught as ApiError).message).toContain("GET /files/missing");
  });

  it("throws with body=null when the error shape does not match AppError", () => {
    let caught: unknown;
    try {
      unwrap({ error: "upstream HTML", response: fakeResponse(502) }, "GET /x");
    } catch (e) {
      caught = e;
    }

    expect(caught).toBeInstanceOf(ApiError);
    expect((caught as ApiError).body).toBeNull();
    expect((caught as ApiError).status).toBe(502);
  });

  it("throws even when error is undefined but data is missing (defensive)", () => {
    expect(() =>
      unwrap({ response: fakeResponse(500) }, "GET /defensive"),
    ).toThrow(ApiError);
  });
});

describe("unwrapVoid", () => {
  it("returns undefined on 2xx with no body", () => {
    expect(
      unwrapVoid({ response: fakeResponse(204) }, "DELETE /files/x"),
    ).toBeUndefined();
  });

  it("throws ApiError with the parsed envelope on error", () => {
    const err: AppError = {
      code: 409,
      status: "conflict",
      message: "already detached",
    };
    expect(() =>
      unwrapVoid(
        { error: err, response: fakeResponse(409) },
        "DELETE /sg/x/files/y",
      ),
    ).toThrow(ApiError);
  });
});

describe("ApiError", () => {
  it("preserves status, response, and body", () => {
    const res = fakeResponse(400);
    const body: AppError = {
      code: 400,
      status: "validation_error",
      message: "name required",
      details: { field: "name" },
    };
    const err = new ApiError("POST /files failed: 400", res, body);

    expect(err.name).toBe("ApiError");
    expect(err.status).toBe(400);
    expect(err.response).toBe(res);
    expect(err.body).toEqual(body);
    expect(err.message).toBe("POST /files failed: 400");
  });

  it("is an Error so existing catch(Error) paths keep working", () => {
    const err = new ApiError("boom", fakeResponse(500), null);
    expect(err instanceof Error).toBe(true);
  });
});
