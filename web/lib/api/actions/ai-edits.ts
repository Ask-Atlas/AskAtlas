/**
 * Server Actions for the `/study-guides/{id}/ai/edits/{edit_id}`
 * audit-row PATCH endpoint (ASK-217). Used by the diff overlay to
 * record whether the user accepted any of the AI's hunks once they
 * resolve a review.
 */
"use server";

import { unwrap } from "../errors";
import { serverApi } from "../server-client";
import type { ApiSchemas } from "../client";

type AIEditAuditRow = ApiSchemas["AIEditAuditRow"];
type UpdateAIEditRequest = ApiSchemas["UpdateAIEditRequest"];

/**
 * Records the user's overall accept/reject decision for a single AI
 * edit audit row. The Go side stamps `accepted_at` server-side, so
 * the client only sends the boolean outcome.
 */
export async function updateAIEdit(
  studyGuideId: string,
  editId: string,
  body: UpdateAIEditRequest,
): Promise<AIEditAuditRow> {
  return unwrap(
    await serverApi.PATCH("/study-guides/{study_guide_id}/ai/edits/{edit_id}", {
      params: {
        path: { study_guide_id: studyGuideId, edit_id: editId },
      },
      body,
    }),
    `PATCH /study-guides/${studyGuideId}/ai/edits/${editId}`,
  );
}
