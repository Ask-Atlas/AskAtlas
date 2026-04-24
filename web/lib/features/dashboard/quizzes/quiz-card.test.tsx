/**
 * Exercises the ASK-175 acceptance criteria: title + question count +
 * updated-at are visible; the Practice CTA targets /practice?quiz={id};
 * and the card as a whole links to /quizzes/{id} (default) or a custom
 * href when provided. Nested-link structure is not directly asserted;
 * it's kept safe by the stretched-link pattern (Practice <Link> is a
 * sibling of the overlay <Link>, not a descendant).
 */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import type { QuizListItemResponse } from "@/lib/api/types";

import { QuizCard } from "./quiz-card";

function makeQuiz(
  overrides: Partial<QuizListItemResponse> = {},
): QuizListItemResponse {
  return {
    id: "q_preview_1",
    title: "CPTS 322 — Midterm Review",
    description: null,
    question_count: 12,
    creator: {
      id: "u_preview_1",
      first_name: "Ada",
      last_name: "Lovelace",
    },
    created_at: "2026-04-18T10:00:00Z",
    updated_at: "2026-04-22T10:00:00Z",
    ...overrides,
  };
}

describe("QuizCard", () => {
  it("shows title, question count, and updated-at (AC1)", () => {
    render(<QuizCard quiz={makeQuiz()} />);
    expect(screen.getByText("CPTS 322 — Midterm Review")).toBeInTheDocument();
    expect(screen.getByText(/12 questions/)).toBeInTheDocument();
    expect(screen.getByText(/Updated/)).toBeInTheDocument();
  });

  it("pluralizes the question count correctly", () => {
    render(<QuizCard quiz={makeQuiz({ question_count: 1 })} />);
    expect(screen.getByText(/1 question/)).toBeInTheDocument();
    expect(screen.queryByText(/1 questions/)).not.toBeInTheDocument();
  });

  it("links the Practice button to /practice?quiz={id} (AC2)", () => {
    render(<QuizCard quiz={makeQuiz()} />);
    expect(screen.getByRole("link", { name: "Practice" })).toHaveAttribute(
      "href",
      "/practice?quiz=q_preview_1",
    );
  });

  it("links the card to /quizzes/{id} by default", () => {
    render(<QuizCard quiz={makeQuiz()} />);
    // The overlay Link is labeled by aria-label ("Open quiz {title}")
    // so queryable by role+name rather than by its invisible text.
    expect(
      screen.getByRole("link", { name: /open quiz cpts 322/i }),
    ).toHaveAttribute("href", "/quizzes/q_preview_1");
  });

  it("honors a custom href on the card", () => {
    render(
      <QuizCard
        quiz={makeQuiz()}
        href="/study-guides/g_1/quizzes/q_preview_1"
      />,
    );
    expect(
      screen.getByRole("link", { name: /open quiz cpts 322/i }),
    ).toHaveAttribute("href", "/study-guides/g_1/quizzes/q_preview_1");
  });

  it("shows the creator's full name", () => {
    render(<QuizCard quiz={makeQuiz()} />);
    expect(screen.getByText(/by Ada Lovelace/)).toBeInTheDocument();
  });
});
