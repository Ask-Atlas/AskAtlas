/**
 * Canonical error type + openapi-fetch result unwrappers.
 *
 * openapi-fetch does NOT throw on 4xx/5xx -- it returns a result tuple
 * `{ data, error, response }` where `error` carries the parsed failure
 * body (typed as `AppError` for every operation in our spec). The
 * helpers here collapse that tuple into the `T | throw` shape actions
 * expect, so callsites read like ordinary `await foo()` instead of
 * the openapi-fetch destructure dance.
 *
 * Keeping the error class here (rather than in `client.ts`) lets both
 * the legacy hand-rolled `apiFetch` wrapper and the new openapi-fetch
 * server actions share one `instanceof ApiError` boundary.
 */
import type { AppError } from "./types";

/**
 * Thrown when the upstream API returns a non-2xx response. Carries
 * the parsed `AppError` envelope when the body was JSON, and the
 * original `Response` so callers can read headers or retry with
 * different auth.
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
 * Shape returned by every `serverApi.GET/POST/PATCH/DELETE(...)` call.
 * Narrowing `data` + `error` with a conditional type would let the
 * compiler know success vs failure without a runtime check, but in
 * practice we always want the explicit branch so the unwrap helpers
 * stay straightforward.
 */
type OpenApiResult<T> = {
  data?: T;
  error?: unknown;
  response: Response;
};

/**
 * Unwrap an openapi-fetch result for an endpoint that returns a JSON
 * body on success. Throws {@link ApiError} when the upstream returned
 * a non-2xx response, preserving the parsed `AppError` envelope.
 *
 * `operation` is a short label like `"GET /me/files"` used in the
 * thrown error message -- callers pass it so stack traces have useful
 * context without a full request dump.
 */
export function unwrap<T>(result: OpenApiResult<T>, operation: string): T {
  if (result.error !== undefined || result.data === undefined) {
    throwApiError(result, operation);
  }
  return result.data;
}

/**
 * Unwrap an openapi-fetch result for an endpoint that returns 204 (no
 * body) on success. Use for DELETE, toggle/record endpoints, and any
 * other void operation. Throws {@link ApiError} on non-2xx.
 */
export function unwrapVoid(
  result: OpenApiResult<unknown>,
  operation: string,
): void {
  if (result.error !== undefined) {
    throwApiError(result, operation);
  }
}

function throwApiError(
  result: OpenApiResult<unknown>,
  operation: string,
): never {
  const body = isAppErrorShape(result.error) ? result.error : null;
  throw new ApiError(
    `${operation} failed: ${result.response.status}`,
    result.response,
    body,
  );
}

function isAppErrorShape(value: unknown): value is AppError {
  if (value === null || typeof value !== "object") return false;
  const candidate = value as Record<string, unknown>;
  return (
    typeof candidate.code === "number" &&
    typeof candidate.status === "string" &&
    typeof candidate.message === "string"
  );
}
