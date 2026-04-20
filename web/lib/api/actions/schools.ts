/**
 * Server Actions for the `/schools` endpoints.
 *
 * Read-only directory of schools, used by the enrollment flow and
 * future course-filter UIs. All calls go through the Clerk-authed
 * `serverApi`.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap } from "../errors";
import type {
  ListSchoolsQuery,
  ListSchoolsResponse,
  SchoolResponse,
} from "../types";

/** List + search schools. Supports keyword search + cursor pagination. */
export async function listSchools(
  query: ListSchoolsQuery = {},
): Promise<ListSchoolsResponse> {
  return unwrap(
    await serverApi.GET("/schools", { params: { query } }),
    "GET /schools",
  );
}

/** Fetch a single school by ID. */
export async function getSchool(schoolId: string): Promise<SchoolResponse> {
  return unwrap(
    await serverApi.GET("/schools/{school_id}", {
      params: { path: { school_id: schoolId } },
    }),
    `GET /schools/${schoolId}`,
  );
}
