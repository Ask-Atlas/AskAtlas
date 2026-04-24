/**
 * Server Actions for the `/courses` endpoints (incl. sections, memberships,
 * and course-scoped study-guide listing/creation).
 *
 * All endpoints require auth -- the Clerk middleware on `serverApi`
 * injects the bearer token per request.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap, unwrapVoid } from "../errors";
import type {
  CourseDetailResponse,
  CourseMemberResponse,
  CreateStudyGuideRequest,
  ListCourseSectionsQuery,
  ListCourseSectionsResponse,
  ListCourseStudyGuidesQuery,
  ListCoursesQuery,
  ListCoursesResponse,
  ListSectionMembersQuery,
  ListSectionMembersResponse,
  ListStudyGuidesResponse,
  MembershipCheckResponse,
  StudyGuideDetailResponse,
} from "../types";

// ---------- Catalogue ----------

/** List + search courses. Supports school filter, department filter, sort, and cursor pagination. */
export async function listCourses(
  query: ListCoursesQuery = {},
): Promise<ListCoursesResponse> {
  return unwrap(
    await serverApi.GET("/courses", { params: { query } }),
    "GET /courses",
  );
}

/** Course detail with embedded section summaries. */
export async function getCourse(
  courseId: string,
): Promise<CourseDetailResponse> {
  return unwrap(
    await serverApi.GET("/courses/{course_id}", {
      params: { path: { course_id: courseId } },
    }),
    `GET /courses/${courseId}`,
  );
}

// ---------- Sections + Membership ----------

/** List sections for a course. Optional `term` filter (e.g. `"fall-2026"`). */
export async function listCourseSections(
  courseId: string,
  query: ListCourseSectionsQuery = {},
): Promise<ListCourseSectionsResponse> {
  return unwrap(
    await serverApi.GET("/courses/{course_id}/sections", {
      params: { path: { course_id: courseId }, query },
    }),
    `GET /courses/${courseId}/sections`,
  );
}

/** List the members of a course section with optional role filter. */
export async function listSectionMembers(
  courseId: string,
  sectionId: string,
  query: ListSectionMembersQuery = {},
): Promise<ListSectionMembersResponse> {
  return unwrap(
    await serverApi.GET("/courses/{course_id}/sections/{section_id}/members", {
      params: {
        path: { course_id: courseId, section_id: sectionId },
        query,
      },
    }),
    `GET /courses/${courseId}/sections/${sectionId}/members`,
  );
}

/** Join a section as the authenticated user. */
export async function joinSection(
  courseId: string,
  sectionId: string,
): Promise<CourseMemberResponse> {
  return unwrap(
    await serverApi.POST("/courses/{course_id}/sections/{section_id}/members", {
      params: { path: { course_id: courseId, section_id: sectionId } },
    }),
    `POST /courses/${courseId}/sections/${sectionId}/members`,
  );
}

/** Check the authenticated user's membership in a section. */
export async function checkMembership(
  courseId: string,
  sectionId: string,
): Promise<MembershipCheckResponse> {
  return unwrap(
    await serverApi.GET(
      "/courses/{course_id}/sections/{section_id}/members/me",
      {
        params: { path: { course_id: courseId, section_id: sectionId } },
      },
    ),
    `GET /courses/${courseId}/sections/${sectionId}/members/me`,
  );
}

/** Leave a section as the authenticated user. */
export async function leaveSection(
  courseId: string,
  sectionId: string,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE(
      "/courses/{course_id}/sections/{section_id}/members/me",
      {
        params: { path: { course_id: courseId, section_id: sectionId } },
      },
    ),
    `DELETE /courses/${courseId}/sections/${sectionId}/members/me`,
  );
}

// ---------- Course-scoped study guides ----------

/** List study guides attached to a course. Supports keyword + tag filter + sort + pagination. */
export async function listCourseStudyGuides(
  courseId: string,
  query: ListCourseStudyGuidesQuery = {},
): Promise<ListStudyGuidesResponse> {
  return unwrap(
    await serverApi.GET("/courses/{course_id}/study-guides", {
      params: { path: { course_id: courseId }, query },
    }),
    `GET /courses/${courseId}/study-guides`,
  );
}

/** Create a study guide attached to a course. */
export async function createStudyGuideForCourse(
  courseId: string,
  body: CreateStudyGuideRequest,
): Promise<StudyGuideDetailResponse> {
  return unwrap(
    await serverApi.POST("/courses/{course_id}/study-guides", {
      params: { path: { course_id: courseId } },
      body,
    }),
    `POST /courses/${courseId}/study-guides`,
  );
}
