import Link from "next/link";
import { type ReactNode } from "react";

import type { CourseResponse } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface CourseCardProps {
  course: CourseResponse;
  variant: "row" | "tile";
  /** Defaults to `/courses/{id}`. Override for contextual routes (e.g. a catalog filter). */
  href?: string;
  /**
   * Trailing affordance: "Joined" badge, FavoriteButton, enrollment CTA, etc.
   * Rendered as a sibling of the stretched Link (not nested inside it) so
   * interactive children can own their own clicks without producing an
   * invalid `<a><button>` HTML structure.
   */
  rightSlot?: ReactNode;
}

export function CourseCard({
  course,
  variant,
  href,
  rightSlot,
}: CourseCardProps) {
  const destination = href ?? `/courses/${course.id}`;
  const code = `${course.department} ${course.number}`;

  return (
    <div
      className={cn(
        "group bg-card focus-within:ring-ring relative rounded-xl border transition-all",
        "hover:shadow-md focus-within:ring-2",
      )}
    >
      {/* Stretched link: invisible overlay that makes the whole card
          clickable without nesting interactive children inside an <a>.
          Content below sits at z-10 so visible text + rightSlot render
          above it; non-interactive content is `pointer-events-none` so
          clicks fall through to the Link, while rightSlot stays
          click-receptive by default. */}
      <Link
        href={destination}
        aria-label={`${code} — ${course.title}`}
        className="absolute inset-0 z-0 rounded-xl focus:outline-none"
      />
      {variant === "row" ? (
        <RowVariant course={course} code={code} rightSlot={rightSlot} />
      ) : (
        <TileVariant course={course} code={code} rightSlot={rightSlot} />
      )}
    </div>
  );
}

function RowVariant({
  course,
  code,
  rightSlot,
}: {
  course: CourseResponse;
  code: string;
  rightSlot?: ReactNode;
}) {
  return (
    <div className="flex items-center gap-3 p-3">
      <div className="pointer-events-none relative z-10 min-w-0 flex-1">
        <p className="text-foreground text-sm font-semibold">{code}</p>
        <p
          className="text-muted-foreground truncate text-xs"
          title={course.title}
        >
          {course.title}
        </p>
      </div>
      {rightSlot ? (
        <div className="relative z-10 shrink-0">{rightSlot}</div>
      ) : null}
    </div>
  );
}

function TileVariant({
  course,
  code,
  rightSlot,
}: {
  course: CourseResponse;
  code: string;
  rightSlot?: ReactNode;
}) {
  return (
    <div className="p-4">
      <div className="pointer-events-none relative z-10 space-y-1.5">
        <h3 className="text-foreground text-base font-semibold leading-tight">
          {code}
        </h3>
        <p className="text-foreground line-clamp-2 text-sm leading-snug">
          {course.title}
        </p>
        <p className="text-muted-foreground truncate text-xs">
          {course.school.name}
        </p>
      </div>
      {rightSlot ? (
        <div className="absolute right-3 top-3 z-10">{rightSlot}</div>
      ) : null}
    </div>
  );
}
