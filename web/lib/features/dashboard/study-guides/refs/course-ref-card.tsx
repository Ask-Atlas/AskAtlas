"use client";

import Link from "next/link";
import { GraduationCap } from "lucide-react";

import { useEntityRef } from "./entity-ref-context";
import { MissingRef } from "./study-guide-ref-card";

interface Props {
  id: string;
  inline?: boolean;
}

export function CourseRefCard({ id, inline }: Props) {
  const { summary, status } = useEntityRef("course", id);

  if (summary == null) {
    return <MissingRef label="Course" inline={inline} status={status} />;
  }

  const href = `/courses/${id}`;
  const code =
    summary.department && summary.number
      ? `${summary.department} ${summary.number}`
      : null;

  if (inline) {
    return (
      <Link
        href={href}
        className="bg-primary/10 text-primary hover:bg-primary/20 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-sm font-medium no-underline"
      >
        <GraduationCap className="size-3.5" aria-hidden />
        {code ?? summary.title ?? "Course"}
      </Link>
    );
  }

  const heading = code && summary.title ? `${code} · ${summary.title}` : code ?? summary.title ?? "Course";

  return (
    <Link
      href={href}
      className="hover:border-primary/40 hover:bg-muted/40 my-3 flex flex-col gap-1 rounded-lg border p-3 no-underline"
    >
      <div className="flex items-center gap-2">
        <GraduationCap className="text-muted-foreground size-4" aria-hidden />
        <span className="text-foreground font-medium">{heading}</span>
      </div>
      {summary.school ? (
        <div className="text-muted-foreground text-xs">
          {summary.school.acronym ? (
            <span className="font-semibold">{summary.school.acronym}</span>
          ) : null}
          {summary.school.acronym && summary.school.name ? (
            <span> · </span>
          ) : null}
          {summary.school.name ? <span>{summary.school.name}</span> : null}
        </div>
      ) : null}
    </Link>
  );
}
