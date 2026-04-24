/**
 * Ergonomic re-exports of the most commonly-used OpenAPI schemas.
 *
 * Generated types live in `./generated/types.ts` and follow the deeply
 * nested `components["schemas"]["..."]` convention, which is noisy at
 * callsites. Callers pull named aliases from here instead -- the
 * definitions stay in sync because they're just re-shaped generics.
 *
 * When a new schema lands in `api/openapi.yaml`:
 *   1. Re-run `make generate-api-types`
 *   2. Add an alias here if any caller will reach for it by name
 *
 * This file is intentionally free of runtime code so it can be imported
 * from both server and browser modules without a `"use server"` or
 * `"server-only"` boundary.
 */
import type { components, paths } from "./generated/types";

// ---------- Namespace re-exports ----------
export type ApiSchemas = components["schemas"];
export type ApiPaths = paths;

// ---------- Files ----------
export type FileResponse = ApiSchemas["FileResponse"];
export type ListFilesResponse = ApiSchemas["ListFilesResponse"];
export type CreateFileRequest = ApiSchemas["CreateFileRequest"];
export type UpdateFileRequest = ApiSchemas["UpdateFileRequest"];
export type CreateGrantRequest = ApiSchemas["CreateGrantRequest"];
export type RevokeGrantRequest = ApiSchemas["RevokeGrantRequest"];
export type GrantResponse = ApiSchemas["GrantResponse"];
export type ToggleFavoriteResponse = ApiSchemas["ToggleFavoriteResponse"];
export type ListFilesQuery = NonNullable<
  ApiPaths["/me/files"]["get"]["parameters"]["query"]
>;

// ---------- Schools ----------
export type SchoolResponse = ApiSchemas["SchoolResponse"];
export type SchoolSummary = ApiSchemas["SchoolSummary"];
export type ListSchoolsResponse = ApiSchemas["ListSchoolsResponse"];
export type ListSchoolsQuery = NonNullable<
  ApiPaths["/schools"]["get"]["parameters"]["query"]
>;

// ---------- Courses ----------
export type CourseResponse = ApiSchemas["CourseResponse"];
export type CourseDetailResponse = ApiSchemas["CourseDetailResponse"];
export type ListCoursesResponse = ApiSchemas["ListCoursesResponse"];
export type ListCoursesQuery = NonNullable<
  ApiPaths["/courses"]["get"]["parameters"]["query"]
>;
export type SectionResponse = ApiSchemas["SectionResponse"];
export type SectionSummary = ApiSchemas["SectionSummary"];
export type ListCourseSectionsResponse =
  ApiSchemas["ListCourseSectionsResponse"];
export type ListCourseSectionsQuery = NonNullable<
  ApiPaths["/courses/{course_id}/sections"]["get"]["parameters"]["query"]
>;
export type CourseMemberResponse = ApiSchemas["CourseMemberResponse"];
export type SectionMemberResponse = ApiSchemas["SectionMemberResponse"];
export type ListSectionMembersResponse =
  ApiSchemas["ListSectionMembersResponse"];
export type ListSectionMembersQuery = NonNullable<
  ApiPaths["/courses/{course_id}/sections/{section_id}/members"]["get"]["parameters"]["query"]
>;
export type MembershipCheckResponse = ApiSchemas["MembershipCheckResponse"];

// ---------- Study Guides ----------
export type StudyGuideDetailResponse = ApiSchemas["StudyGuideDetailResponse"];
export type StudyGuideListItemResponse =
  ApiSchemas["StudyGuideListItemResponse"];
export type ListStudyGuidesResponse = ApiSchemas["ListStudyGuidesResponse"];
export type ListCourseStudyGuidesQuery = NonNullable<
  ApiPaths["/courses/{course_id}/study-guides"]["get"]["parameters"]["query"]
>;
export type CreateStudyGuideRequest = ApiSchemas["CreateStudyGuideRequest"];
export type UpdateStudyGuideRequest = ApiSchemas["UpdateStudyGuideRequest"];
export type AttachResourceRequest = ApiSchemas["AttachResourceRequest"];
export type ResourceSummary = ApiSchemas["ResourceSummary"];
export type FileAttachmentResponse = ApiSchemas["FileAttachmentResponse"];
export type StudyGuideFileSummary = ApiSchemas["StudyGuideFileSummary"];
export type RecommendationResponse = ApiSchemas["RecommendationResponse"];
export type CastVoteRequest = ApiSchemas["CastVoteRequest"];
export type CastVoteResponse = ApiSchemas["CastVoteResponse"];
// ASK-211/ASK-212: visibility + grant sharing
export type StudyGuideVisibility = NonNullable<
  ApiSchemas["StudyGuideDetailResponse"]["visibility"]
>;
export type StudyGuideCreateGrantRequest =
  ApiSchemas["StudyGuideCreateGrantRequest"];
export type StudyGuideRevokeGrantRequest =
  ApiSchemas["StudyGuideRevokeGrantRequest"];
export type StudyGuideGrantResponse = ApiSchemas["StudyGuideGrantResponse"];
export type ListStudyGuideGrantsResponse =
  ApiSchemas["ListStudyGuideGrantsResponse"];

// ---------- Quizzes ----------
export type QuizDetailResponse = ApiSchemas["QuizDetailResponse"];
export type QuizListItemResponse = ApiSchemas["QuizListItemResponse"];
export type ListQuizzesResponse = ApiSchemas["ListQuizzesResponse"];
export type CreateQuizRequest = ApiSchemas["CreateQuizRequest"];
export type UpdateQuizRequest = ApiSchemas["UpdateQuizRequest"];
export type CreateQuizQuestion = ApiSchemas["CreateQuizQuestion"];
export type CreateQuizMCQOption = ApiSchemas["CreateQuizMCQOption"];
export type QuizQuestionResponse = ApiSchemas["QuizQuestionResponse"];

// ---------- Practice Sessions ----------
export type PracticeSessionResponse = ApiSchemas["PracticeSessionResponse"];
export type SessionDetailResponse = ApiSchemas["SessionDetailResponse"];
export type SessionSummaryResponse = ApiSchemas["SessionSummaryResponse"];
export type ListSessionsResponse = ApiSchemas["ListSessionsResponse"];
export type ListPracticeSessionsQuery = NonNullable<
  ApiPaths["/quizzes/{quiz_id}/sessions"]["get"]["parameters"]["query"]
>;
export type CompletedSessionResponse = ApiSchemas["CompletedSessionResponse"];
export type SubmitAnswerRequest = ApiSchemas["SubmitAnswerRequest"];
export type PracticeAnswerResponse = ApiSchemas["PracticeAnswerResponse"];

// ---------- Me / Dashboard ----------
export type DashboardResponse = ApiSchemas["DashboardResponse"];
export type ListMyEnrollmentsResponse = ApiSchemas["ListMyEnrollmentsResponse"];
export type ListMyEnrollmentsQuery = NonNullable<
  ApiPaths["/me/courses"]["get"]["parameters"]["query"]
>;
export type ListMyStudyGuidesResponse = ApiSchemas["ListMyStudyGuidesResponse"];
export type ListMyStudyGuidesQuery = NonNullable<
  ApiPaths["/me/study-guides"]["get"]["parameters"]["query"]
>;
export type ListFavoritesResponse = ApiSchemas["ListFavoritesResponse"];
export type ListFavoritesQuery = NonNullable<
  ApiPaths["/me/favorites"]["get"]["parameters"]["query"]
>;
export type ListRecentsResponse = ApiSchemas["ListRecentsResponse"];
export type ListRecentsQuery = NonNullable<
  ApiPaths["/me/recents"]["get"]["parameters"]["query"]
>;
export type FavoriteItem = ApiSchemas["FavoriteItem"];
export type RecentItem = ApiSchemas["RecentItem"];

// ---------- Refs (ASK-208 / ASK-209 inline entity cards) ----------
export type RefsResolveRequest = ApiSchemas["RefsResolveRequest"];
export type RefsResolveRequestRef = ApiSchemas["RefsResolveRequestRef"];
export type RefsResolveResponse = ApiSchemas["RefsResolveResponse"];
export type RefSummary = ApiSchemas["RefSummary"];
export type RefCourseInfo = ApiSchemas["RefCourseInfo"];
export type RefCreatorInfo = ApiSchemas["RefCreatorInfo"];
export type RefSchoolInfo = ApiSchemas["RefSchoolInfo"];

// ---------- Errors ----------
export type AppError = ApiSchemas["AppError"];
