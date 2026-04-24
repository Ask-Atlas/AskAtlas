import Link from "next/link";

import { Button } from "@/components/ui/button";
import type { QuizListItemResponse } from "@/lib/api/types";
import { cn, formatRelativeDate } from "@/lib/utils";

interface QuizCardProps {
  quiz: QuizListItemResponse;
  /** Defaults to `/quizzes/{id}`. Override for contextual routes. */
  href?: string;
}

export function QuizCard({ quiz, href }: QuizCardProps) {
  const destination = href ?? `/quizzes/${quiz.id}`;
  const creatorName =
    `${quiz.creator.first_name} ${quiz.creator.last_name}`.trim();
  const questionLabel =
    quiz.question_count === 1
      ? "1 question"
      : `${quiz.question_count} questions`;

  return (
    <div
      className={cn(
        "group bg-card focus-within:ring-ring relative rounded-xl border transition-all",
        "hover:shadow-md focus-within:ring-2",
      )}
    >
      {/* Stretched link: the whole card navigates to the quiz detail,
          but the Practice button below is a sibling (not a descendant)
          of this Link so it can own its own navigation without
          producing an invalid nested <a><a> tree. */}
      <Link
        href={destination}
        aria-label={`Open quiz ${quiz.title}`}
        className="absolute inset-0 z-0 rounded-xl focus:outline-none"
      />
      {/* Content layer is pointer-events-none so clicks anywhere on the
          card (padding, gap, text) pass through to the overlay Link.
          Only the Practice button re-enables pointer events so it can
          receive its own navigation click. */}
      <div className="pointer-events-none flex items-center gap-3 p-4">
        <div className="relative z-10 min-w-0 flex-1">
          <p
            className="text-foreground line-clamp-2 text-sm font-semibold leading-snug"
            title={quiz.title}
          >
            {quiz.title}
          </p>
          <p className="text-muted-foreground mt-1 text-xs">
            {questionLabel}
            <span className="mx-1.5">·</span>
            <span>by {creatorName}</span>
            <span className="mx-1.5">·</span>
            <span>Updated {formatRelativeDate(quiz.updated_at)}</span>
          </p>
        </div>
        <div className="pointer-events-auto relative z-10 shrink-0">
          <Button asChild size="sm">
            <Link href={`/practice?quiz=${quiz.id}`}>Practice</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
