/**
 * Tests for the practice page wire-up (ASK-192).
 *
 * The page is a client component; we exercise it with RTL through the
 * full session lifecycle (start -> submit -> complete) and the empty /
 * missing-quiz states. `next/navigation` and `@/lib/api` are mocked at
 * the module level so the test owns every API + routing call.
 */
import { act, fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import type {
  CompletedSessionResponse,
  PracticeAnswerResponse,
  PracticeSessionResponse,
  QuizDetailResponse,
} from "@/lib/api/types";

const pushMock = jest.fn();

jest.mock("next/navigation", () => ({
  useSearchParams: jest.fn(),
  useRouter: () => ({ push: pushMock, replace: jest.fn(), refresh: jest.fn() }),
}));

jest.mock("../../lib/api", () => ({
  getQuiz: jest.fn(),
  startPracticeSession: jest.fn(),
  submitPracticeAnswer: jest.fn(),
  completePracticeSession: jest.fn(),
  abandonPracticeSession: jest.fn(),
}));

jest.mock("../../lib/features/shared/toast/toast", () => ({
  toast: { error: jest.fn(), success: jest.fn(), info: jest.fn() },
}));

import { useSearchParams } from "next/navigation";

import {
  abandonPracticeSession,
  completePracticeSession,
  getQuiz,
  startPracticeSession,
  submitPracticeAnswer,
} from "../../lib/api";

import PracticePage from "./page";

const useSearchParamsMock = useSearchParams as jest.MockedFunction<
  typeof useSearchParams
>;
const getQuizMock = getQuiz as jest.MockedFunction<typeof getQuiz>;
const startMock = startPracticeSession as jest.MockedFunction<
  typeof startPracticeSession
>;
const submitMock = submitPracticeAnswer as jest.MockedFunction<
  typeof submitPracticeAnswer
>;
const completeMock = completePracticeSession as jest.MockedFunction<
  typeof completePracticeSession
>;
const abandonMock = abandonPracticeSession as jest.MockedFunction<
  typeof abandonPracticeSession
>;

function setQuizQuery(quizId: string | null) {
  useSearchParamsMock.mockReturnValue({
    get: (key: string) => (key === "quiz" ? quizId : null),
  } as unknown as ReturnType<typeof useSearchParams>);
}

function makeQuiz(
  overrides: Partial<QuizDetailResponse> = {},
): QuizDetailResponse {
  return {
    id: "q_1",
    study_guide_id: "sg_1",
    title: "Spring Quiz",
    description: "A short quiz.",
    creator: { id: "u_1", first_name: "Ada", last_name: "Lovelace" },
    questions: [
      {
        id: "qid_1",
        type: "multiple-choice",
        question: "Capital of France?",
        options: ["London", "Paris", "Berlin"],
        correct_answer: "Paris",
        hint: "Eiffel Tower lives there.",
        feedback: { correct: "Nice!", incorrect: "It's Paris." },
        sort_order: 1,
      },
      {
        id: "qid_2",
        type: "true-false",
        question: "The sky is green.",
        correct_answer: false,
        hint: null,
        feedback: { correct: "Right.", incorrect: "It's blue." },
        sort_order: 2,
      },
    ],
    created_at: "2026-04-20T10:00:00Z",
    updated_at: "2026-04-20T10:00:00Z",
    ...overrides,
  };
}

function makeSession(
  overrides: Partial<PracticeSessionResponse> = {},
): PracticeSessionResponse {
  return {
    id: "s_1",
    quiz_id: "q_1",
    started_at: "2026-04-27T10:00:00Z",
    completed_at: null,
    total_questions: 2,
    correct_answers: 0,
    answers: [],
    ...overrides,
  };
}

function makeAnswer(
  overrides: Partial<PracticeAnswerResponse> = {},
): PracticeAnswerResponse {
  return {
    question_id: "qid_1",
    user_answer: "Paris",
    is_correct: true,
    verified: true,
    answered_at: "2026-04-27T10:00:30Z",
    ...overrides,
  };
}

function makeCompleted(
  overrides: Partial<CompletedSessionResponse> = {},
): CompletedSessionResponse {
  return {
    id: "s_1",
    quiz_id: "q_1",
    started_at: "2026-04-27T10:00:00Z",
    completed_at: "2026-04-27T10:05:00Z",
    total_questions: 2,
    correct_answers: 2,
    score_percentage: 100,
    ...overrides,
  };
}

async function flush() {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
}

beforeEach(() => {
  jest.clearAllMocks();
});

describe("PracticePage (ASK-192)", () => {
  it("shows the no-quiz nudge when ?quiz= is missing", () => {
    setQuizQuery(null);

    render(<PracticePage />);

    expect(screen.getByText("Pick a quiz to practice")).toBeInTheDocument();
    expect(getQuizMock).not.toHaveBeenCalled();
    expect(startMock).not.toHaveBeenCalled();
  });

  it("renders the empty-quiz state when the quiz has zero questions", async () => {
    setQuizQuery("q_1");
    getQuizMock.mockResolvedValue(makeQuiz({ questions: [] }));
    startMock.mockResolvedValue(makeSession({ total_questions: 0 }));

    render(<PracticePage />);
    await flush();

    expect(
      screen.getByText("This quiz has no questions yet"),
    ).toBeInTheDocument();
  });

  it("walks the full happy path: start -> submit -> complete", async () => {
    setQuizQuery("q_1");
    getQuizMock.mockResolvedValue(makeQuiz());
    startMock.mockResolvedValue(makeSession());
    submitMock.mockResolvedValueOnce(makeAnswer()).mockResolvedValueOnce(
      makeAnswer({
        question_id: "qid_2",
        user_answer: "false",
        is_correct: true,
      }),
    );
    completeMock.mockResolvedValue(makeCompleted());

    render(<PracticePage />);
    await flush();

    expect(screen.getByText("Capital of France?")).toBeInTheDocument();
    expect(screen.getByText("1 / 2")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Paris" }));
    fireEvent.click(screen.getByRole("button", { name: /submit answer/i }));
    await flush();

    expect(submitMock).toHaveBeenCalledWith("s_1", {
      question_id: "qid_1",
      user_answer: "Paris",
    });
    expect(screen.getByText("Correct!")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /next question/i }));
    await flush();

    expect(screen.getByText("The sky is green.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "False" }));
    fireEvent.click(screen.getByRole("button", { name: /submit answer/i }));
    await flush();

    expect(submitMock).toHaveBeenLastCalledWith("s_1", {
      question_id: "qid_2",
      user_answer: "false",
    });

    fireEvent.click(screen.getByRole("button", { name: /finish practice/i }));
    await flush();

    expect(completeMock).toHaveBeenCalledWith("s_1");
    expect(screen.getByText("Practice Complete")).toBeInTheDocument();
    expect(screen.getByText("100%")).toBeInTheDocument();
  });

  it("resumes mid-session by jumping to the first unanswered question", async () => {
    setQuizQuery("q_1");
    getQuizMock.mockResolvedValue(makeQuiz());
    startMock.mockResolvedValue(
      makeSession({ correct_answers: 1, answers: [makeAnswer()] }),
    );

    render(<PracticePage />);
    await flush();

    // qid_1 is already answered; we should land on qid_2.
    expect(screen.getByText("The sky is green.")).toBeInTheDocument();
    expect(screen.getByText("2 / 2")).toBeInTheDocument();
  });

  it("auto-completes when resuming a session with every answer already in", async () => {
    setQuizQuery("q_1");
    getQuizMock.mockResolvedValue(makeQuiz());
    startMock.mockResolvedValue(
      makeSession({
        correct_answers: 2,
        answers: [
          makeAnswer(),
          makeAnswer({
            question_id: "qid_2",
            user_answer: "false",
            is_correct: true,
          }),
        ],
      }),
    );
    completeMock.mockResolvedValue(makeCompleted());

    render(<PracticePage />);
    await flush();

    expect(completeMock).toHaveBeenCalledWith("s_1");
    expect(screen.getByText("Practice Complete")).toBeInTheDocument();
  });

  it("abandons the session when Leave is confirmed", async () => {
    setQuizQuery("q_1");
    getQuizMock.mockResolvedValue(makeQuiz());
    startMock.mockResolvedValue(makeSession());
    abandonMock.mockResolvedValue(undefined);

    render(<PracticePage />);
    await flush();

    fireEvent.click(screen.getByRole("button", { name: /leave/i }));
    const confirmButtons = screen.getAllByRole("button", { name: /^leave$/i });
    // The dialog's confirm button is the last one rendered.
    fireEvent.click(confirmButtons[confirmButtons.length - 1]);
    await flush();

    expect(abandonMock).toHaveBeenCalledWith("s_1");
    expect(pushMock).toHaveBeenCalledWith("/home");
  });
});
