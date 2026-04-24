/**
 * Server Actions for the `/me/files` + `/files/*` endpoints.
 *
 * Every export is a Next.js Server Action: callable from Server
 * Components directly and from Client Components over RPC. The
 * `"use server"` directive at the top of the file turns every exported
 * async function into a remote procedure automatically.
 *
 * Auth is handled by the Clerk middleware wired into `serverApi`, so
 * these functions never reach for a token themselves.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap, unwrapVoid } from "../errors";
import type {
  CreateFileRequest,
  CreateGrantRequest,
  FileResponse,
  GrantResponse,
  ListFilesQuery,
  ListFilesResponse,
  RevokeGrantRequest,
  ToggleFavoriteResponse,
  UpdateFileRequest,
} from "../types";

/** List files for the current user (supports filters, pagination, search). */
export async function listFiles(
  query: ListFilesQuery = {},
): Promise<ListFilesResponse> {
  return unwrap(
    await serverApi.GET("/me/files", { params: { query } }),
    "GET /me/files",
  );
}

/**
 * Create a file metadata record in `pending` status. The Next.js
 * server is expected to generate the S3 key first and pass it as
 * `s3_key`; a subsequent PATCH transitions the record to `complete`
 * or `failed`.
 */
export async function createFile(
  body: CreateFileRequest,
): Promise<FileResponse> {
  return unwrap(await serverApi.POST("/files", { body }), "POST /files");
}

/** Get a file record by ID. */
export async function getFile(fileId: string): Promise<FileResponse> {
  return unwrap(
    await serverApi.GET("/files/{file_id}", {
      params: { path: { file_id: fileId } },
    }),
    `GET /files/${fileId}`,
  );
}

/**
 * Patch file metadata and/or transition upload status. Both request
 * fields are optional but at least one must be present.
 */
export async function updateFile(
  fileId: string,
  body: UpdateFileRequest,
): Promise<FileResponse> {
  return unwrap(
    await serverApi.PATCH("/files/{file_id}", {
      params: { path: { file_id: fileId } },
      body,
    }),
    `PATCH /files/${fileId}`,
  );
}

/** Soft-delete a file. */
export async function deleteFile(fileId: string): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/files/{file_id}", {
      params: { path: { file_id: fileId } },
    }),
    `DELETE /files/${fileId}`,
  );
}

/**
 * Record that the authenticated user opened/previewed this file.
 * Fire-and-forget; response is 204.
 */
export async function recordFileView(fileId: string): Promise<void> {
  return unwrapVoid(
    await serverApi.POST("/files/{file_id}/view", {
      params: { path: { file_id: fileId } },
    }),
    `POST /files/${fileId}/view`,
  );
}

/** Toggle the favorite flag on a file (idempotent per final state). */
export async function toggleFileFavorite(
  fileId: string,
): Promise<ToggleFavoriteResponse> {
  return unwrap(
    await serverApi.POST("/files/{file_id}/favorite", {
      params: { path: { file_id: fileId } },
    }),
    `POST /files/${fileId}/favorite`,
  );
}

/** Grant a permission on a file to another user or course section. */
export async function createGrant(
  fileId: string,
  body: CreateGrantRequest,
): Promise<GrantResponse> {
  return unwrap(
    await serverApi.POST("/files/{file_id}/grants", {
      params: { path: { file_id: fileId } },
      body,
    }),
    `POST /files/${fileId}/grants`,
  );
}

/** Revoke a previously-issued file permission grant. */
export async function revokeGrant(
  fileId: string,
  body: RevokeGrantRequest,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/files/{file_id}/grants", {
      params: { path: { file_id: fileId } },
      body,
    }),
    `DELETE /files/${fileId}/grants`,
  );
}
