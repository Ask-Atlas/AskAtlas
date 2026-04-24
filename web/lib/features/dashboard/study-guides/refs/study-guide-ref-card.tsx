"use client";

import Link from "next/link";
import { BookOpen, Sparkles } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

import { useEntityRef } from "./entity-ref-context";

interface Props {
  id: string;
  inline?: boolean;
}

export function StudyGuideRefCard({ id, inline }: Props) {
  const { summary, status } = useEntityRef("sg", id);

  if (summary == null) {
    return <MissingRef label="Study guide" inline={inline} status={status} />;
  }

  const href = `/study-guides/${id}`;
  const courseLabel = summary.course
    ? `${summary.course.department} ${summary.course.number}`
    : null;

  if (inline) {
    return (
      <Link
        href={href}
        className="bg-primary/10 text-primary hover:bg-primary/20 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-sm font-medium no-underline"
      >
        <BookOpen className="size-3.5" aria-hidden />
        {summary.title ?? "Study guide"}
      </Link>
    );
  }

  return (
    <Link
      href={href}
      className="hover:border-primary/40 hover:bg-muted/40 my-3 flex flex-col gap-1 rounded-lg border p-3 no-underline"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-2">
          <BookOpen className="text-muted-foreground size-4" aria-hidden />
          <span className="text-foreground font-medium">
            {summary.title ?? "Study guide"}
          </span>
        </div>
        {summary.is_recommended ? (
          <Badge variant="secondary" className="gap-1 text-xs">
            <Sparkles className="size-3" aria-hidden />
            Recommended
          </Badge>
        ) : null}
      </div>
      <div className="text-muted-foreground flex items-center gap-3 text-xs">
        {courseLabel ? <span>{courseLabel}</span> : null}
        {typeof summary.quiz_count === "number" ? (
          <span>
            {summary.quiz_count}{" "}
            {summary.quiz_count === 1 ? "quiz" : "quizzes"}
          </span>
        ) : null}
      </div>
    </Link>
  );
}

interface MissingRefProps {
  label: string;
  inline?: boolean;
  status: "pending" | "ready";
}

export function MissingRef({ label, inline, status }: MissingRefProps) {
  const copy =
    status === "pending"
      ? `Loading ${label.toLowerCase()}…`
      : `${label} unavailable`;
  if (inline) {
    return (
      <span
        className={cn(
          "bg-muted text-muted-foreground inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-sm",
          status === "pending" && "animate-pulse",
        )}
      >
        {copy}
      </span>
    );
  }
  return (
    <div
      className={cn(
        "bg-muted/40 text-muted-foreground my-3 rounded-lg border border-dashed p-3 text-sm",
        status === "pending" && "animate-pulse",
      )}
    >
      {copy}
    </div>
  );
}
