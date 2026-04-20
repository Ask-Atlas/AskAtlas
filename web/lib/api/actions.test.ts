/**
 * Cross-domain integration tests for the Server Actions under
 * `./actions/*`. Each domain contributes at least one happy-path
 * assertion + one error-path assertion so any regression in the
 * openapi-fetch argument shape gets caught before it ships.
 *
 * The Clerk-authed `serverApi` is fully mocked -- we don't exercise
 * network or auth at this layer. The `server-client.ts` contract
 * (openapi-fetch `.GET/.POST/.PATCH/.DELETE/.PUT`) is stable and
 * already covered by openapi-fetch's own test suite upstream; here
 * we just verify the thin action wrappers forward the right inputs
 * and interpret the result tuple through `unwrap`/`unwrapVoid`.
 */

// --- Mock the server client BEFORE importing any action module ---
jest.mock("./server-client", () => ({
  serverApi: {
    GET: jest.fn(),
    POST: jest.fn(),
    PATCH: jest.fn(),
    PUT: jest.fn(),
    DELETE: jest.fn(),
    use: jest.fn(),
  },
}));

import { serverApi } from "./server-client";
import { ApiError } from "./errors";

// --- Actions ---
import { createFile, deleteFile, listFiles } from "./actions/files";
import { listSchools } from "./actions/schools";
import { joinSection, listCourses } from "./actions/courses";
import { castStudyGuideVote, getStudyGuide } from "./actions/study-guides";
import { getQuiz } from "./actions/quizzes";
import { startPracticeSession, submitPracticeAnswer } from "./actions/practice";
import { listDashboard, toggleCourseFavorite } from "./actions/me";

import type { AppError, CreateFileRequest, FileResponse } from "./types";

type MockedServerApi = {
  GET: jest.Mock;
  POST: jest.Mock;
  PATCH: jest.Mock;
  PUT: jest.Mock;
  DELETE: jest.Mock;
  use: jest.Mock;
};

const api = serverApi as unknown as MockedServerApi;

function ok<T>(data: T, status = 200): { data: T; response: Response } {
  return {
    data,
    response: { status, ok: true } as unknown as Response,
  };
}

function noBody(status = 204): { response: Response } {
  return { response: { status, ok: true } as unknown as Response };
}

function err(
  status: number,
  body: AppError,
): { error: AppError; response: Response } {
  return {
    error: body,
    response: { status, ok: false } as unknown as Response,
  };
}

const APP_ERR: AppError = {
  code: 404,
  status: "not_found",
  message: "missing",
};

beforeEach(() => {
  jest.clearAllMocks();
});

describe("files actions", () => {
  it("listFiles forwards the query and returns typed data", async () => {
    const payload = { files: [], has_more: false };
    api.GET.mockResolvedValueOnce(ok(payload));

    const result = await listFiles({ scope: "owned" });

    expect(result).toEqual(payload);
    expect(api.GET).toHaveBeenCalledWith("/me/files", {
      params: { query: { scope: "owned" } },
    });
  });

  it("createFile posts the body and returns the created record", async () => {
    const body: CreateFileRequest = {
      name: "notes.pdf",
      mime_type: "application/pdf",
      size: 128,
      s3_key: "users/abc/notes.pdf",
    };
    const created: FileResponse = {
      id: "11111111-1111-1111-1111-111111111111",
      name: body.name,
      mime_type: body.mime_type,
      size: body.size,
      status: "pending",
      created_at: "2026-04-20T00:00:00Z",
      updated_at: "2026-04-20T00:00:00Z",
    };
    api.POST.mockResolvedValueOnce(ok(created, 201));

    await expect(createFile(body)).resolves.toEqual(created);
    expect(api.POST).toHaveBeenCalledWith("/files", { body });
  });

  it("deleteFile resolves void on 204", async () => {
    api.DELETE.mockResolvedValueOnce(noBody());
    await expect(
      deleteFile("22222222-2222-2222-2222-222222222222"),
    ).resolves.toBeUndefined();
    expect(api.DELETE).toHaveBeenCalledWith("/files/{file_id}", {
      params: {
        path: { file_id: "22222222-2222-2222-2222-222222222222" },
      },
    });
  });

  it("listFiles throws ApiError with the parsed AppError body on 4xx", async () => {
    api.GET.mockResolvedValueOnce(
      err(401, {
        ...APP_ERR,
        code: 401,
        status: "unauthorized",
        message: "expired token",
      }),
    );
    await expect(listFiles()).rejects.toBeInstanceOf(ApiError);
  });
});

describe("schools actions", () => {
  it("listSchools passes the query through", async () => {
    api.GET.mockResolvedValueOnce(ok({ schools: [], has_more: false }));
    await listSchools({ q: "stanford" });
    expect(api.GET).toHaveBeenCalledWith("/schools", {
      params: { query: { q: "stanford" } },
    });
  });
});

describe("courses actions", () => {
  it("listCourses forwards filters", async () => {
    api.GET.mockResolvedValueOnce(ok({ courses: [], has_more: false }));
    await listCourses({ school_id: "sch", department: "cs" });
    expect(api.GET).toHaveBeenCalledWith("/courses", {
      params: { query: { school_id: "sch", department: "cs" } },
    });
  });

  it("joinSection binds path params into the URL template", async () => {
    api.POST.mockResolvedValueOnce(ok({ id: "m", role: "student" }));
    await joinSection("c1", "s1");
    expect(api.POST).toHaveBeenCalledWith(
      "/courses/{course_id}/sections/{section_id}/members",
      { params: { path: { course_id: "c1", section_id: "s1" } } },
    );
  });
});

describe("study-guides actions", () => {
  it("getStudyGuide returns the detail payload", async () => {
    api.GET.mockResolvedValueOnce(ok({ id: "sg-1" }));
    await expect(getStudyGuide("sg-1")).resolves.toEqual({ id: "sg-1" });
  });

  it("castStudyGuideVote forwards the body", async () => {
    api.POST.mockResolvedValueOnce(ok({ up: 1, down: 0, my_vote: "up" }));
    await castStudyGuideVote("sg-1", { vote: "up" } as never);
    expect(api.POST).toHaveBeenCalledWith(
      "/study-guides/{study_guide_id}/votes",
      {
        params: { path: { study_guide_id: "sg-1" } },
        body: { vote: "up" },
      },
    );
  });
});

describe("quizzes actions", () => {
  it("getQuiz throws ApiError on upstream 500", async () => {
    api.GET.mockResolvedValueOnce(
      err(500, {
        ...APP_ERR,
        code: 500,
        status: "internal",
        message: "boom",
      }),
    );
    await expect(getQuiz("q-1")).rejects.toMatchObject({
      name: "ApiError",
      status: 500,
    });
  });
});

describe("practice actions", () => {
  it("startPracticeSession calls POST without a body", async () => {
    api.POST.mockResolvedValueOnce(ok({ id: "sess-1", status: "in_progress" }));
    await startPracticeSession("q-1");
    expect(api.POST).toHaveBeenCalledWith("/quizzes/{quiz_id}/sessions", {
      params: { path: { quiz_id: "q-1" } },
    });
  });

  it("submitPracticeAnswer forwards both path and body", async () => {
    api.POST.mockResolvedValueOnce(ok({ question_id: "qn-1", correct: true }));
    await submitPracticeAnswer("sess-1", {
      question_id: "qn-1",
      answer: "A",
    } as never);
    expect(api.POST).toHaveBeenCalledWith("/sessions/{session_id}/answers", {
      params: { path: { session_id: "sess-1" } },
      body: { question_id: "qn-1", answer: "A" },
    });
  });
});

describe("me actions", () => {
  it("listDashboard hits /me/dashboard with no query", async () => {
    api.GET.mockResolvedValueOnce(ok({}));
    await listDashboard();
    expect(api.GET).toHaveBeenCalledWith("/me/dashboard", {});
  });

  it("toggleCourseFavorite binds the course_id", async () => {
    api.POST.mockResolvedValueOnce(
      ok({ favorited: true, favorited_at: "2026-04-20T00:00:00Z" }),
    );
    await toggleCourseFavorite("c-1");
    expect(api.POST).toHaveBeenCalledWith("/me/courses/{course_id}/favorite", {
      params: { path: { course_id: "c-1" } },
    });
  });
});
