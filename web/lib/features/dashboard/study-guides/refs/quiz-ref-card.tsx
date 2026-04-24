"use client";

import Link from "next/link";
import { Play, Target } from "lucide-react";

import { Button } from "@/components/ui/button";

import { useEntityRef } from "./entity-ref-context";
import { MissingRef } from "./study-guide-ref-card";

interface Props {
  id: string;
  inline?: boolean;
}

export function QuizRefCard({ id, inline }: Props) {
  const { summary, status } = useEntityRef("quiz", id);

  if (summary == null) {
    return <MissingRef label="Quiz" inline={inline} status={status} />;
  }

  const practiceHref = `/practice?quiz=${id}`;

  if (inline) {
    return (
      <Link
        href={practiceHref}
        className="bg-primary/10 text-primary hover:bg-primary/20 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-sm font-medium no-underline"
      >
        <Target className="size-3.5" aria-hidden />
        {summary.title ?? "Quiz"}
      </Link>
    );
  }

  const creatorName = summary.creator
    ? `${summary.creator.first_name} ${summary.creator.last_name}`.trim()
    : null;

  return (
    <div className="my-3 flex items-center justify-between gap-3 rounded-lg border p-3">
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-2">
          <Target className="text-muted-foreground size-4" aria-hidden />
          <span className="text-foreground font-medium">
            {summary.title ?? "Quiz"}
          </span>
        </div>
        <div className="text-muted-foreground flex items-center gap-3 text-xs">
          {typeof summary.question_count === "number" ? (
            <span>
              {summary.question_count}{" "}
              {summary.question_count === 1 ? "question" : "questions"}
            </span>
          ) : null}
          {creatorName ? <span>by {creatorName}</span> : null}
        </div>
      </div>
      <Button asChild size="sm" className="shrink-0">
        <Link href={practiceHref}>
          <Play className="size-4" aria-hidden />
          Practice
        </Link>
      </Button>
    </div>
  );
}
