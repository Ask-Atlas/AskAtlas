/**
 * Server Actions for the practice-session endpoints.
 *
 * Session lifecycle:
 *   1. POST /quizzes/{id}/sessions          -> start (or resume in-progress)
 *   2. POST /sessions/{id}/answers          -> submit one answer per question
 *   3. POST /sessions/{id}/complete         -> finalize + return score
 *      or DELETE /sessions/{id}             -> abandon without scoring
 *
 * Listing + detail support history UIs and deep-linking into previous
 * attempts.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap, unwrapVoid } from "../errors";
import type {
  CompletedSessionResponse,
  ListPracticeSessionsQuery,
  ListSessionsResponse,
  PracticeAnswerResponse,
  PracticeSessionResponse,
  SessionDetailResponse,
  SubmitAnswerRequest,
} from "../types";

/** List the caller's practice sessions for a quiz. Supports status filter + pagination. */
export async function listPracticeSessions(
  quizId: string,
  query: ListPracticeSessionsQuery = {},
): Promise<ListSessionsResponse> {
  return unwrap(
    await serverApi.GET("/quizzes/{quiz_id}/sessions", {
      params: { path: { quiz_id: quizId }, query },
    }),
    `GET /quizzes/${quizId}/sessions`,
  );
}

/**
 * Start a new practice session or resume an existing in-progress one
 * for the given quiz. The API picks between 200 (resume) and 201
 * (fresh session) transparently.
 */
export async function startPracticeSession(
  quizId: string,
): Promise<PracticeSessionResponse> {
  return unwrap(
    await serverApi.POST("/quizzes/{quiz_id}/sessions", {
      params: { path: { quiz_id: quizId } },
    }),
    `POST /quizzes/${quizId}/sessions`,
  );
}

/** Fetch a practice session detail including all submitted answers. */
export async function getPracticeSession(
  sessionId: string,
): Promise<SessionDetailResponse> {
  return unwrap(
    await serverApi.GET("/sessions/{session_id}", {
      params: { path: { session_id: sessionId } },
    }),
    `GET /sessions/${sessionId}`,
  );
}

/** Hard-delete an in-progress practice session (no scoring). */
export async function abandonPracticeSession(sessionId: string): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/sessions/{session_id}", {
      params: { path: { session_id: sessionId } },
    }),
    `DELETE /sessions/${sessionId}`,
  );
}

/** Mark a practice session as completed and return the score. */
export async function completePracticeSession(
  sessionId: string,
): Promise<CompletedSessionResponse> {
  return unwrap(
    await serverApi.POST("/sessions/{session_id}/complete", {
      params: { path: { session_id: sessionId } },
    }),
    `POST /sessions/${sessionId}/complete`,
  );
}

/** Submit an answer for a single question in an active practice session. */
export async function submitPracticeAnswer(
  sessionId: string,
  body: SubmitAnswerRequest,
): Promise<PracticeAnswerResponse> {
  return unwrap(
    await serverApi.POST("/sessions/{session_id}/answers", {
      params: { path: { session_id: sessionId } },
      body,
    }),
    `POST /sessions/${sessionId}/answers`,
  );
}
