/**
 * Thin fetch wrapper keyed to the generated OpenAPI types (ASK-118).
 *
 * The Go API contract lives in `api/openapi.yaml`. Types under
 * `./generated/types.ts` are produced by `pnpm generate:api-types`
 * (or `make types`) and must not be hand-edited.
 *
 * This file intentionally stays small: it re-exports the handful of
 * aliases callers reach for most often plus one thin `apiFetch` helper
 * so later tickets (ASK-109/112/126/135/146) can wire up individual
 * endpoints without repeating the fetch + error plumbing.
 */
import type { components, paths } from "./generated/types";

/** All named schemas from the OpenAPI spec. */
export type ApiSchemas = components["schemas"];

/** All path operations from the OpenAPI spec. */
export type ApiPaths = paths;

export type FileResponse = ApiSchemas["FileResponse"];
export type ListFilesResponse = ApiSchemas["ListFilesResponse"];
export type AppError = ApiSchemas["AppError"];

/**
 * Base URL for API calls. Falls back to `/api` so same-origin fetches
 * flow through whatever proxy/rewrite the Next.js app configures.
 * Overridable via `NEXT_PUBLIC_API_BASE_URL` when the frontend needs
 * to talk to the Go service on a different origin.
 */
export const API_BASE: string = process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";

/**
 * Error thrown by {@link apiFetch} when the upstream response is not
 * a 2xx. The original `Response` is kept so callers can inspect
 * headers or re-read the body; `body` holds the already-parsed
 * {@link AppError} envelope when the response contained JSON.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly response: Response;
  readonly body: AppError | null;

  constructor(message: string, response: Response, body: AppError | null) {
    super(message);
    this.name = "ApiError";
    this.status = response.status;
    this.response = response;
    this.body = body;
  }
}

/**
 * Typed fetch wrapper that returns the decoded JSON body as `T`.
 * Callers pick `T` from the generated `components["schemas"]` aliases
 * (e.g. `FileResponse`). Non-2xx responses throw {@link ApiError}
 * with the parsed `AppError` envelope attached when available.
 */
export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const url = path.startsWith("http") ? path : `${API_BASE}${path}`;
  const res = await fetch(url, init);

  if (!res.ok) {
    let body: AppError | null = null;
    try {
      body = (await res.clone().json()) as AppError;
    } catch {
      body = null;
    }
    throw new ApiError(
      `${init?.method ?? "GET"} ${url} failed: ${res.status}`,
      res,
      body,
    );
  }

  return (await res.json()) as T;
}

/**
 * Example typed endpoint binding. Follow-ups (ASK-109/112) will add
 * similar helpers per resource; this one exists so downstream files
 * have a shape to copy and so the generated types have a live
 * reference that the TypeScript compiler must resolve.
 */
export function getFile(fileId: string): Promise<FileResponse> {
  return apiFetch<FileResponse>(`/files/${fileId}`);
}
