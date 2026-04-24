/**
 * Server Actions for the `/quizzes/*` endpoints (incl. question CRUD).
 *
 * Listing + creation are keyed off a parent study guide; the rest
 * operate on quiz IDs directly.
 */
"use server";

import { serverApi } from "../server-client";
import { unwrap, unwrapVoid } from "../errors";
import type {
  CreateQuizQuestion,
  CreateQuizRequest,
  ListQuizzesResponse,
  QuizDetailResponse,
  QuizQuestionResponse,
  UpdateQuizRequest,
} from "../types";

// ---------- Study-guide-scoped ----------

/** List quizzes attached to a study guide. */
export async function listQuizzes(
  studyGuideId: string,
): Promise<ListQuizzesResponse> {
  return unwrap(
    await serverApi.GET("/study-guides/{study_guide_id}/quizzes", {
      params: { path: { study_guide_id: studyGuideId } },
    }),
    `GET /study-guides/${studyGuideId}/quizzes`,
  );
}

/** Create a quiz attached to a study guide (questions optional at create time). */
export async function createQuiz(
  studyGuideId: string,
  body: CreateQuizRequest,
): Promise<QuizDetailResponse> {
  return unwrap(
    await serverApi.POST("/study-guides/{study_guide_id}/quizzes", {
      params: { path: { study_guide_id: studyGuideId } },
      body,
    }),
    `POST /study-guides/${studyGuideId}/quizzes`,
  );
}

// ---------- Quiz CRUD ----------

/** Fetch a quiz with all questions and correct answers. */
export async function getQuiz(quizId: string): Promise<QuizDetailResponse> {
  return unwrap(
    await serverApi.GET("/quizzes/{quiz_id}", {
      params: { path: { quiz_id: quizId } },
    }),
    `GET /quizzes/${quizId}`,
  );
}

/** Update a quiz's metadata (title / description). */
export async function updateQuiz(
  quizId: string,
  body: UpdateQuizRequest,
): Promise<QuizDetailResponse> {
  return unwrap(
    await serverApi.PATCH("/quizzes/{quiz_id}", {
      params: { path: { quiz_id: quizId } },
      body,
    }),
    `PATCH /quizzes/${quizId}`,
  );
}

/** Soft-delete a quiz. */
export async function deleteQuiz(quizId: string): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/quizzes/{quiz_id}", {
      params: { path: { quiz_id: quizId } },
    }),
    `DELETE /quizzes/${quizId}`,
  );
}

// ---------- Question CRUD ----------

/** Append a new question to a quiz. */
export async function addQuizQuestion(
  quizId: string,
  body: CreateQuizQuestion,
): Promise<QuizQuestionResponse> {
  return unwrap(
    await serverApi.POST("/quizzes/{quiz_id}/questions", {
      params: { path: { quiz_id: quizId } },
      body,
    }),
    `POST /quizzes/${quizId}/questions`,
  );
}

/** Replace a question in a quiz (PUT semantics -- full replacement). */
export async function replaceQuizQuestion(
  quizId: string,
  questionId: string,
  body: CreateQuizQuestion,
): Promise<QuizQuestionResponse> {
  return unwrap(
    await serverApi.PUT("/quizzes/{quiz_id}/questions/{question_id}", {
      params: { path: { quiz_id: quizId, question_id: questionId } },
      body,
    }),
    `PUT /quizzes/${quizId}/questions/${questionId}`,
  );
}

/** Delete a question from a quiz. */
export async function deleteQuizQuestion(
  quizId: string,
  questionId: string,
): Promise<void> {
  return unwrapVoid(
    await serverApi.DELETE("/quizzes/{quiz_id}/questions/{question_id}", {
      params: { path: { quiz_id: quizId, question_id: questionId } },
    }),
    `DELETE /quizzes/${quizId}/questions/${questionId}`,
  );
}
