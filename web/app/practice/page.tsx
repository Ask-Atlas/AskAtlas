/**
 * Practice page wire-up (ASK-192).
 *
 * Reads `?quiz={id}`, fetches the quiz detail (for questions) and a
 * practice session (created or resumed) in parallel, and walks the
 * user through one question at a time. The server is the source of
 * truth for `is_correct`; we display feedback from the submit
 * response, never from a local string compare.
 *
 * Phases:
 *   - `no-quiz`     no `?quiz=` query param; nudge to /study-guides
 *   - `loading`     initial parallel fetch in flight
 *   - `error`       fetch failed; allow retry
 *   - `empty-quiz`  quiz has zero questions
 *   - `playing`     answering questions
 *   - `completed`   show ScoreScreen after completePracticeSession
 *
 * Resume: if the session already has answers, jump to the first
 * unanswered question (by id). Already-submitted answers are kept in
 * a `Map<questionId, PracticeAnswerResponse>` so revisiting a finished
 * question would still render its outcome -- though the UX advances
 * straight to the next unanswered slot on load.
 */
"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  abandonPracticeSession,
  completePracticeSession,
  getQuiz,
  startPracticeSession,
  submitPracticeAnswer,
} from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import type {
  CompletedSessionResponse,
  PracticeAnswerResponse,
  PracticeSessionResponse,
  QuizDetailResponse,
  QuizQuestionResponse,
} from "@/lib/api/types";
import { ConfirmationDialog } from "@/lib/features/shared/confirmation-dialog";
import { toast } from "@/lib/features/shared/toast/toast";
import { cn } from "@/lib/utils";

type Phase =
  | { kind: "no-quiz" }
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "empty-quiz"; quiz: QuizDetailResponse }
  | {
      kind: "playing";
      quiz: QuizDetailResponse;
      session: PracticeSessionResponse;
      currentIndex: number;
      answersByQuestionId: Map<string, PracticeAnswerResponse>;
    }
  | { kind: "completed"; result: CompletedSessionResponse };

export default function PracticePage() {
  // Next.js requires `useSearchParams` to live under a Suspense boundary
  // so static-export can bail out cleanly when the URL is unknown at
  // build time.
  return (
    <Suspense fallback={<LoadingState />}>
      <PracticePageInner />
    </Suspense>
  );
}

function PracticePageInner() {
  const params = useSearchParams();
  const quizId = params.get("quiz");

  const [phase, setPhase] = useState<Phase>(
    quizId ? { kind: "loading" } : { kind: "no-quiz" },
  );

  useEffect(() => {
    if (!quizId) return;
    let cancelled = false;

    Promise.all([getQuiz(quizId), startPracticeSession(quizId)])
      .then(async ([quiz, session]) => {
        if (cancelled) return;
        if (quiz.questions.length === 0) {
          setPhase({ kind: "empty-quiz", quiz });
          return;
        }
        const answersByQuestionId = answersByQuestionMap(session.answers);
        const allAnswered = quiz.questions.every((q) =>
          answersByQuestionId.has(q.id),
        );
        // A resumed session with every question already answered (or
        // one the API returns with `completed_at` set) skips the
        // player and finalizes immediately. Calling complete is
        // idempotent on a finished session.
        if (allAnswered || session.completed_at !== null) {
          const result = await completePracticeSession(session.id);
          if (cancelled) return;
          setPhase({ kind: "completed", result });
          return;
        }
        const startIndex = quiz.questions.findIndex(
          (q) => !answersByQuestionId.has(q.id),
        );
        setPhase({
          kind: "playing",
          quiz,
          session,
          currentIndex: startIndex,
          answersByQuestionId,
        });
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        toast.error(err);
        setPhase({
          kind: "error",
          message:
            err instanceof ApiError
              ? `Couldn't start practice (${err.status}).`
              : "Couldn't start practice.",
        });
      });

    return () => {
      cancelled = true;
    };
  }, [quizId]);

  if (phase.kind === "no-quiz") return <NoQuizState />;
  if (phase.kind === "loading") return <LoadingState />;
  if (phase.kind === "error") return <ErrorState message={phase.message} />;
  if (phase.kind === "empty-quiz") return <EmptyQuizState quiz={phase.quiz} />;
  if (phase.kind === "completed") return <ScoreScreen result={phase.result} />;

  return <Player phase={phase} setPhase={setPhase} />;
}

function answersByQuestionMap(
  answers: PracticeAnswerResponse[],
): Map<string, PracticeAnswerResponse> {
  const m = new Map<string, PracticeAnswerResponse>();
  for (const a of answers) {
    if (a.question_id) m.set(a.question_id, a);
  }
  return m;
}

interface PlayerProps {
  phase: Extract<Phase, { kind: "playing" }>;
  setPhase: (p: Phase) => void;
}

function Player({ phase, setPhase }: PlayerProps) {
  const router = useRouter();
  const { quiz, session, currentIndex, answersByQuestionId } = phase;
  const currentQuestion = quiz.questions[currentIndex];
  const existingAnswer = currentQuestion
    ? answersByQuestionId.get(currentQuestion.id)
    : undefined;

  const [draftAnswer, setDraftAnswer] = useState<string | null>(
    existingAnswer?.user_answer ?? null,
  );
  const [submitting, setSubmitting] = useState(false);
  const [showHint, setShowHint] = useState(false);
  const [leaving, setLeaving] = useState(false);
  const [completing, setCompleting] = useState(false);
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false);

  // Reset transient UI when the active question changes. We deliberately
  // depend only on currentIndex so a fresh submit response (which mutates
  // existingAnswer) doesn't blow away the draft mid-render.
  useEffect(() => {
    setDraftAnswer(existingAnswer?.user_answer ?? null);
    setShowHint(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentIndex]);

  const total = quiz.questions.length;
  const isLast = currentIndex === total - 1;
  const showFeedback = Boolean(existingAnswer);

  async function onSubmit() {
    if (!currentQuestion || draftAnswer === null || draftAnswer === "") return;
    setSubmitting(true);
    try {
      const answer = await submitPracticeAnswer(session.id, {
        question_id: currentQuestion.id,
        user_answer: draftAnswer,
      });
      const next = new Map(answersByQuestionId);
      next.set(currentQuestion.id, answer);
      setPhase({ ...phase, answersByQuestionId: next });
    } catch (err: unknown) {
      toast.error(err);
    } finally {
      setSubmitting(false);
    }
  }

  function onNext() {
    if (currentIndex < total - 1) {
      setPhase({ ...phase, currentIndex: currentIndex + 1 });
    }
  }

  async function onComplete() {
    setCompleting(true);
    try {
      const result = await completePracticeSession(session.id);
      setPhase({ kind: "completed", result });
    } catch (err: unknown) {
      toast.error(err);
    } finally {
      setCompleting(false);
    }
  }

  async function onLeaveConfirm() {
    setLeaving(true);
    try {
      await abandonPracticeSession(session.id);
      setShowLeaveConfirm(false);
      router.push("/home");
    } catch (err: unknown) {
      toast.error(err);
    } finally {
      setLeaving(false);
    }
  }

  if (!currentQuestion) return null;

  return (
    <div className="min-h-screen bg-black text-white">
      <div className="border-b border-white/10 bg-white/5">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-4">
            <Button
              onClick={() => setShowLeaveConfirm(true)}
              variant="ghost"
              size="sm"
              className="text-gray-400 hover:text-white"
            >
              ← Leave
            </Button>
            <div>
              <h2 className="font-semibold">{quiz.title}</h2>
              {quiz.description ? (
                <p className="text-sm text-gray-400">{quiz.description}</p>
              ) : null}
            </div>
          </div>
          <div className="text-right">
            <p className="text-sm text-gray-400">Progress</p>
            <p className="text-lg font-semibold">
              {currentIndex + 1} / {total}
            </p>
          </div>
        </div>
      </div>

      <div className="mx-auto max-w-4xl px-6 py-12">
        <div className="rounded-xl border border-white/10 bg-white/5 p-8">
          <Badge
            className={cn(
              "mb-6",
              currentQuestion.type === "multiple-choice" &&
                "border-blue-500/50 text-blue-400",
              currentQuestion.type === "true-false" &&
                "border-green-500/50 text-green-400",
              currentQuestion.type === "freeform" &&
                "border-purple-500/50 text-purple-400",
            )}
            variant="outline"
          >
            {labelForType(currentQuestion.type)}
          </Badge>

          <h3 className="mb-6 text-xl font-semibold">
            {currentQuestion.question}
          </h3>

          {currentQuestion.type === "multiple-choice" &&
            currentQuestion.options && (
              <div className="mb-6 space-y-3">
                {currentQuestion.options.map((option) => (
                  <button
                    key={option}
                    type="button"
                    onClick={() => setDraftAnswer(option)}
                    disabled={showFeedback || submitting}
                    className={cn(
                      "w-full rounded-xl border-2 p-4 text-left transition-all duration-200",
                      draftAnswer === option
                        ? "border-orange-500 bg-orange-500/10"
                        : "border-white/10 hover:border-orange-500/50 hover:bg-orange-500/5",
                    )}
                  >
                    {option}
                  </button>
                ))}
              </div>
            )}

          {currentQuestion.type === "true-false" && (
            <div className="mb-6 grid grid-cols-2 gap-4">
              {[
                { label: "True", value: "true" },
                { label: "False", value: "false" },
              ].map(({ label, value }) => (
                <button
                  key={value}
                  type="button"
                  onClick={() => setDraftAnswer(value)}
                  disabled={showFeedback || submitting}
                  className={cn(
                    "rounded-xl border-2 p-6 text-lg font-semibold transition-all duration-200",
                    draftAnswer === value
                      ? "border-orange-500 bg-orange-500/10"
                      : "border-white/10 hover:border-orange-500/50 hover:bg-orange-500/5",
                  )}
                >
                  {label}
                </button>
              ))}
            </div>
          )}

          {currentQuestion.type === "freeform" && (
            <div className="mb-6">
              <Input
                type="text"
                placeholder="Type your answer here..."
                value={draftAnswer ?? ""}
                onChange={(e) => setDraftAnswer(e.target.value)}
                disabled={showFeedback || submitting}
                className="w-full rounded-xl border-2 border-white/10 bg-white/5 p-4 text-lg focus:border-orange-500"
              />
            </div>
          )}

          {!showFeedback && currentQuestion.hint ? (
            <div className="mb-6">
              <Button
                onClick={() => setShowHint((v) => !v)}
                variant="outline"
                size="sm"
                className="border-yellow-500/30 text-yellow-500 hover:bg-yellow-500/10"
              >
                {showHint ? "Hide Hint" : "💡 Show Hint"}
              </Button>
              {showHint ? (
                <div className="mt-4 rounded-lg border border-yellow-500/20 bg-yellow-500/5 p-4">
                  <p className="text-sm text-yellow-200">
                    {currentQuestion.hint}
                  </p>
                </div>
              ) : null}
            </div>
          ) : null}

          {existingAnswer ? (
            <FeedbackBlock
              question={currentQuestion}
              answer={existingAnswer}
              isLast={isLast}
              completing={completing}
              onNext={onNext}
              onComplete={onComplete}
            />
          ) : (
            <Button
              onClick={onSubmit}
              disabled={
                submitting || draftAnswer === null || draftAnswer === ""
              }
              className="w-full bg-orange-500 text-white hover:bg-orange-600"
            >
              {submitting ? "Submitting..." : "Submit Answer"}
            </Button>
          )}
        </div>
      </div>

      <ConfirmationDialog
        open={showLeaveConfirm}
        onOpenChange={setShowLeaveConfirm}
        title="Leave practice?"
        description="Your in-progress session will be discarded."
        confirmLabel={leaving ? "Leaving..." : "Leave"}
        cancelLabel="Stay"
        destructive
        disabled={leaving}
        onConfirm={onLeaveConfirm}
      />
    </div>
  );
}

function FeedbackBlock({
  question,
  answer,
  isLast,
  completing,
  onNext,
  onComplete,
}: {
  question: QuizQuestionResponse;
  answer: PracticeAnswerResponse;
  isLast: boolean;
  completing: boolean;
  onNext: () => void;
  onComplete: () => void;
}) {
  const isCorrect = answer.is_correct === true;
  const message = isCorrect
    ? question.feedback.correct
    : question.feedback.incorrect;

  return (
    <>
      <div
        className={cn(
          "mb-6 rounded-xl border-2 p-6",
          isCorrect
            ? "border-green-500/50 bg-green-500/10"
            : "border-red-500/50 bg-red-500/10",
        )}
      >
        <div className="flex items-start gap-4">
          <div className="text-3xl">{isCorrect ? "🎉" : "📚"}</div>
          <div>
            <h3
              className={cn(
                "mb-2 text-lg font-semibold",
                isCorrect ? "text-green-400" : "text-red-400",
              )}
            >
              {isCorrect ? "Correct!" : "Not quite"}
            </h3>
            {message ? <p className="text-gray-300">{message}</p> : null}
            {!answer.verified ? (
              <p className="mt-2 text-xs text-gray-400">
                Freeform answers are scored by string match — your wording may
                be acceptable even if marked otherwise.
              </p>
            ) : null}
          </div>
        </div>
      </div>

      {isLast ? (
        <Button
          onClick={onComplete}
          disabled={completing}
          className="w-full bg-orange-500 text-white hover:bg-orange-600"
        >
          {completing ? "Finishing..." : "Finish Practice"}
        </Button>
      ) : (
        <Button
          onClick={onNext}
          className="w-full bg-orange-500 text-white hover:bg-orange-600"
        >
          Next Question →
        </Button>
      )}
    </>
  );
}

function ScoreScreen({ result }: { result: CompletedSessionResponse }) {
  return (
    <div className="min-h-screen bg-black text-white">
      <div className="mx-auto max-w-2xl px-6 py-20 text-center">
        <Badge className="mb-6 border-orange-500/20 bg-orange-500/10 text-orange-500">
          Practice Complete
        </Badge>
        <h1 className="mb-4 text-5xl font-bold">
          {Math.round(result.score_percentage)}%
        </h1>
        <p className="mb-8 text-lg text-gray-400">
          You got {result.correct_answers} of {result.total_questions} correct.
        </p>
        <div className="flex justify-center gap-3">
          <Button asChild variant="outline">
            <Link href="/home">Back home</Link>
          </Button>
          <Button
            asChild
            className="bg-orange-500 text-white hover:bg-orange-600"
          >
            <Link href="/me/study-guides">My study guides</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}

function NoQuizState() {
  return (
    <CenteredMessage
      title="Pick a quiz to practice"
      body="Open a study guide and start a quiz from there."
      cta={{ label: "Browse my study guides", href: "/me/study-guides" }}
    />
  );
}

function EmptyQuizState({ quiz }: { quiz: QuizDetailResponse }) {
  return (
    <CenteredMessage
      title="This quiz has no questions yet"
      body={`"${quiz.title}" doesn't have any questions to practice.`}
      cta={{
        label: "Open the study guide",
        href: `/study-guides/${quiz.study_guide_id}`,
      }}
    />
  );
}

function LoadingState() {
  return (
    <div className="min-h-screen bg-black text-white">
      <div className="mx-auto max-w-2xl px-6 py-20 text-center text-gray-400">
        Loading practice session…
      </div>
    </div>
  );
}

function ErrorState({ message }: { message: string }) {
  return (
    <CenteredMessage
      title="Practice unavailable"
      body={message}
      cta={{ label: "Back home", href: "/home" }}
    />
  );
}

function CenteredMessage({
  title,
  body,
  cta,
}: {
  title: string;
  body: string;
  cta: { label: string; href: string };
}) {
  return (
    <div className="min-h-screen bg-black text-white">
      <div className="mx-auto max-w-2xl px-6 py-20 text-center">
        <h1 className="mb-3 text-3xl font-semibold">{title}</h1>
        <p className="mb-8 text-gray-400">{body}</p>
        <Button
          asChild
          className="bg-orange-500 text-white hover:bg-orange-600"
        >
          <Link href={cta.href}>{cta.label}</Link>
        </Button>
      </div>
    </div>
  );
}

function labelForType(type: QuizQuestionResponse["type"]): string {
  if (type === "multiple-choice") return "Multiple Choice";
  if (type === "true-false") return "True / False";
  return "Free Response";
}
