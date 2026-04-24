/**
 * Exercises the OpenAPI-generated types + thin fetch wrapper (ASK-118).
 *
 * Primary goal: confirm the types resolve end-to-end. If codegen regresses
 * (missing schema, renamed field, broken $ref) TypeScript blocks CI here
 * before the runtime assertions even run. The runtime checks are small --
 * just enough to pin the wrapper's happy path + AppError envelope parsing.
 */
import type { components, paths } from "./generated/types";
import {
  API_BASE,
  ApiError,
  apiFetch,
  getFile,
  type AppError,
  type FileResponse,
  type ListFilesResponse,
} from "./client";

describe("lib/api generated types", () => {
  it("resolves core schemas from the OpenAPI spec", () => {
    const file: FileResponse = {
      id: "00000000-0000-0000-0000-000000000000",
      name: "notes.pdf",
      size: 1024,
      mime_type: "application/pdf",
      status: "complete",
      created_at: "2026-04-20T00:00:00Z",
      updated_at: "2026-04-20T00:00:00Z",
      favorited_at: null,
      last_viewed_at: null,
    };

    const list: ListFilesResponse = { files: [file], has_more: false };
    expect(list.files[0].id).toBe(file.id);
  });

  it("preserves AppError shape from the spec", () => {
    const err: AppError = {
      code: 400,
      status: "validation_error",
      message: "size must be positive",
      details: { field: "size" },
    };
    expect(err.code).toBe(400);
    expect(err.details?.field).toBe("size");
  });

  it("exposes the paths tree for per-operation typing", () => {
    // Compile-only: if the path is renamed or the operation signature
    // changes, the alias below will stop resolving.
    type ListFilesOp = paths["/me/files"]["get"];
    type ListFilesQuery = NonNullable<ListFilesOp["parameters"]["query"]>;
    const query: ListFilesQuery = {};
    expect(query).toEqual({});
  });

  it("surfaces every declared schema under components.schemas", () => {
    // Spot-check a handful the downstream tickets will reach for.
    type StudyGuide = components["schemas"]["StudyGuideDetailResponse"];
    type Course = components["schemas"]["CourseResponse"];
    type Quiz = components["schemas"]["QuizDetailResponse"];
    const sample: Pick<StudyGuide, "id"> &
      Pick<Course, "id"> &
      Pick<Quiz, "id"> = { id: "x" };
    expect(sample.id).toBe("x");
  });
});

/**
 * Minimal Response-like stub. The jsdom test env does not expose the
 * global `Response` constructor, so we hand-roll the surface apiFetch
 * touches (`ok`, `status`, `clone`, `json`). `jsonBody` is the parsed
 * payload for the happy path; pass `null` with `textBody` set to
 * simulate a non-JSON response.
 */
function makeResponse(
  status: number,
  jsonBody: unknown,
  opts: { textBody?: string } = {},
): Response {
  const ok = status >= 200 && status < 300;
  const base = {
    ok,
    status,
    clone() {
      return this;
    },
    async json() {
      if (opts.textBody !== undefined) throw new SyntaxError("not JSON");
      return jsonBody;
    },
  };
  return base as unknown as Response;
}

describe("apiFetch", () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("returns the decoded body on 2xx", async () => {
    const payload: FileResponse = {
      id: "11111111-1111-1111-1111-111111111111",
      name: "ok.pdf",
      size: 1,
      mime_type: "application/pdf",
      status: "complete",
      created_at: "2026-04-20T00:00:00Z",
      updated_at: "2026-04-20T00:00:00Z",
    };
    global.fetch = jest.fn().mockResolvedValue(makeResponse(200, payload));

    const result = await apiFetch<FileResponse>("/files/11111111");
    expect(result).toEqual(payload);
    expect(global.fetch).toHaveBeenCalledWith(
      `${API_BASE}/files/11111111`,
      undefined,
    );
  });

  it("throws ApiError with parsed AppError body on 4xx", async () => {
    const errBody: AppError = {
      code: 404,
      status: "not_found",
      message: "file missing",
    };
    global.fetch = jest.fn().mockResolvedValue(makeResponse(404, errBody));

    await expect(getFile("missing-id")).rejects.toMatchObject({
      name: "ApiError",
      status: 404,
      body: errBody,
    });
  });

  it("keeps ApiError.body null when the response is not JSON", async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValue(
        makeResponse(502, null, { textBody: "upstream blew up" }),
      );

    const caught = await apiFetch<FileResponse>("/files/broken").catch(
      (e: unknown) => e,
    );
    expect(caught).toBeInstanceOf(ApiError);
    expect((caught as ApiError).status).toBe(502);
    expect((caught as ApiError).body).toBeNull();
  });
});
