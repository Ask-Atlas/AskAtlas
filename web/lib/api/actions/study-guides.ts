/**
 * Server Actions for the `/study-guides/*` endpoints.
 *
 * Covers CRUD, resource + file attach/detach, recommendations, and
 * voting. Creation lives under `courses.ts` (`createStudyGuideForCourse`)
 * because the API only creates study guides in the context of a parent
 * course.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap, unwrapVoid } from "../errors";
import type {
  AttachResourceRequest,
  CastVoteRequest,
  CastVoteResponse,
  FileAttachmentResponse,
  ListStudyGuideGrantsResponse,
  RecommendationResponse,
  ResourceSummary,
  StudyGuideCreateGrantRequest,
  StudyGuideDetailResponse,
  StudyGuideGrantResponse,
  StudyGuideRevokeGrantRequest,
  UpdateStudyGuideRequest,
} from "../types";

// ---------- CRUD ----------

/** Fetch a study guide detail. */
export async function getStudyGuide(
  studyGuideId: string,
): Promise<StudyGuideDetailResponse> {
  return unwrap(
    await serverApi.GET("/study-guides/{study_guide_id}", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `GET /study-guides/${studyGuideId}`,
  );
}

/** Patch a study guide's editable fields (title, article, tags, etc.). */
export async function updateStudyGuide(
  studyGuideId: string,
  body: UpdateStudyGuideRequest,
): Promise<StudyGuideDetailResponse> {
  return unwrap(
    await serverApi.PATCH("/study-guides/{study_guide_id}", {
      params: { path: { study_guide_id: studyGuideId } },
      body,
    }),
    `PATCH /study-guides/${studyGuideId}`,
  );
}

/** Soft-delete a study guide. */
export async function deleteStudyGuide(studyGuideId: string): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/study-guides/{study_guide_id}", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `DELETE /study-guides/${studyGuideId}`,
  );
}

// ---------- Resource attachments ----------

/** Attach an external resource (URL/citation) to a study guide. */
export async function attachResource(
  studyGuideId: string,
  body: AttachResourceRequest,
): Promise<ResourceSummary> {
  return unwrap(
    await serverApi.POST("/study-guides/{study_guide_id}/resources", {
      params: { path: { study_guide_id: studyGuideId } },
      body,
    }),
    `POST /study-guides/${studyGuideId}/resources`,
  );
}

/** Detach a resource from a study guide. */
export async function detachResource(
  studyGuideId: string,
  resourceId: string,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE(
      "/study-guides/{study_guide_id}/resources/{resource_id}",
      {
        params: {
          path: { study_guide_id: studyGuideId, resource_id: resourceId },
        },
      },
    ),
    `DELETE /study-guides/${studyGuideId}/resources/${resourceId}`,
  );
}

// ---------- File attachments ----------

/** Attach a pre-existing file record to a study guide. */
export async function attachFile(
  studyGuideId: string,
  fileId: string,
): Promise<FileAttachmentResponse> {
  return unwrap(
    await serverApi.POST("/study-guides/{study_guide_id}/files/{file_id}", {
      params: { path: { study_guide_id: studyGuideId, file_id: fileId } },
    }),
    `POST /study-guides/${studyGuideId}/files/${fileId}`,
  );
}

/** Detach a file from a study guide. */
export async function detachFile(
  studyGuideId: string,
  fileId: string,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/study-guides/{study_guide_id}/files/{file_id}", {
      params: { path: { study_guide_id: studyGuideId, file_id: fileId } },
    }),
    `DELETE /study-guides/${studyGuideId}/files/${fileId}`,
  );
}

// ---------- Recommendations ----------

/** Recommend a study guide (adds the caller to its recommender list). */
export async function recommendStudyGuide(
  studyGuideId: string,
): Promise<RecommendationResponse> {
  return unwrap(
    await serverApi.POST("/study-guides/{study_guide_id}/recommendations", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `POST /study-guides/${studyGuideId}/recommendations`,
  );
}

/** Remove the caller's recommendation on a study guide. */
export async function removeStudyGuideRecommendation(
  studyGuideId: string,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/study-guides/{study_guide_id}/recommendations", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `DELETE /study-guides/${studyGuideId}/recommendations`,
  );
}

// ---------- Votes ----------

/** Cast (or change) an up/down vote on a study guide. */
export async function castStudyGuideVote(
  studyGuideId: string,
  body: CastVoteRequest,
): Promise<CastVoteResponse> {
  return unwrap(
    await serverApi.POST("/study-guides/{study_guide_id}/votes", {
      params: { path: { study_guide_id: studyGuideId } },
      body,
    }),
    `POST /study-guides/${studyGuideId}/votes`,
  );
}

/** Remove the caller's vote on a study guide. */
export async function removeStudyGuideVote(
  studyGuideId: string,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/study-guides/{study_guide_id}/votes", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `DELETE /study-guides/${studyGuideId}/votes`,
  );
}

// ---------- Grants (ASK-211/ASK-212) ----------

/** List the grants (share targets) currently attached to a study guide. */
export async function listStudyGuideGrants(
  studyGuideId: string,
): Promise<ListStudyGuideGrantsResponse> {
  return unwrap(
    await serverApi.GET("/study-guides/{study_guide_id}/grants", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `GET /study-guides/${studyGuideId}/grants`,
  );
}

/** Grant a permission on a study guide to another user or course. */
export async function createStudyGuideGrant(
  studyGuideId: string,
  body: StudyGuideCreateGrantRequest,
): Promise<StudyGuideGrantResponse> {
  return unwrap(
    await serverApi.POST("/study-guides/{study_guide_id}/grants", {
      params: { path: { study_guide_id: studyGuideId } },
      body,
    }),
    `POST /study-guides/${studyGuideId}/grants`,
  );
}

/** Revoke a previously-issued study-guide permission grant. */
export async function revokeStudyGuideGrant(
  studyGuideId: string,
  body: StudyGuideRevokeGrantRequest,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/study-guides/{study_guide_id}/grants", {
      params: { path: { study_guide_id: studyGuideId } },
      body,
    }),
    `DELETE /study-guides/${studyGuideId}/grants`,
  );
}
