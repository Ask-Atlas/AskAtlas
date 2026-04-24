/**
 * Server Actions for the `/me/*` endpoints (dashboard, enrollments,
 * my study guides, favorites, recents, favorite toggles).
 *
 * These power the authenticated home page, sidebar nav, and "my"
 * filters across the app. Everything here requires an active Clerk
 * session; the auth middleware on `serverApi` injects the token.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap } from "../errors";
import type {
  DashboardResponse,
  ListFavoritesQuery,
  ListFavoritesResponse,
  ListMyEnrollmentsQuery,
  ListMyEnrollmentsResponse,
  ListMyStudyGuidesQuery,
  ListMyStudyGuidesResponse,
  ListRecentsQuery,
  ListRecentsResponse,
  ToggleFavoriteResponse,
} from "../types";

/** Aggregated dashboard payload: courses, study guides, practice, files. */
export async function listDashboard(): Promise<DashboardResponse> {
  return unwrap(await serverApi.GET("/me/dashboard", {}), "GET /me/dashboard");
}

/** Caller's section enrollments. Optional `term` + `role` filter. */
export async function listMyEnrollments(
  query: ListMyEnrollmentsQuery = {},
): Promise<ListMyEnrollmentsResponse> {
  return unwrap(
    await serverApi.GET("/me/courses", { params: { query } }),
    "GET /me/courses",
  );
}

/** Study guides authored by the caller. */
export async function listMyStudyGuides(
  query: ListMyStudyGuidesQuery = {},
): Promise<ListMyStudyGuidesResponse> {
  return unwrap(
    await serverApi.GET("/me/study-guides", { params: { query } }),
    "GET /me/study-guides",
  );
}

/** Caller's favorited items across files, study guides, and courses. */
export async function listFavorites(
  query: ListFavoritesQuery = {},
): Promise<ListFavoritesResponse> {
  return unwrap(
    await serverApi.GET("/me/favorites", { params: { query } }),
    "GET /me/favorites",
  );
}

/** Caller's most recently viewed items (files, study guides, courses). */
export async function listRecents(
  query: ListRecentsQuery = {},
): Promise<ListRecentsResponse> {
  return unwrap(
    await serverApi.GET("/me/recents", { params: { query } }),
    "GET /me/recents",
  );
}

/** Toggle the study-guide favorite flag for the caller. */
export async function toggleStudyGuideFavorite(
  studyGuideId: string,
): Promise<ToggleFavoriteResponse> {
  return unwrap(
    await serverApi.POST("/me/study-guides/{study_guide_id}/favorite", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `POST /me/study-guides/${studyGuideId}/favorite`,
  );
}

/** Toggle the course favorite flag for the caller. */
export async function toggleCourseFavorite(
  courseId: string,
): Promise<ToggleFavoriteResponse> {
  return unwrap(
    await serverApi.POST("/me/courses/{course_id}/favorite", {
      params: { path: { course_id: courseId } },
    }),
    `POST /me/courses/${courseId}/favorite`,
  );
}
