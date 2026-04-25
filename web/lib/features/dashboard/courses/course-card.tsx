import Link from "next/link";
import { type ReactNode } from "react";
import { ArrowRight, Layers, Users } from "lucide-react";

import { Skeleton } from "@/components/ui/skeleton";
import type { CourseResponse } from "@/lib/api/types";
import { cn } from "@/lib/utils";

// Shared shell so the skeleton renders in the exact same outer
// shape as the real card. Any change to padding / radius / border
// here flows to both the populated card and the loading state, so
// the two never drift.
const TILE_SHELL_CLASSES =
  "group bg-card focus-within:ring-ring relative rounded-[10px] border border-zinc-200 transition-all";
const TILE_INNER_CLASSES =
  "flex min-h-[180px] flex-col gap-2.5 px-5 pb-4 pt-[18px]";

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
  /** Tile-only: aggregate enrolled members across all sections. */
  memberCount?: number;
  /** Tile-only: number of sections offered for this course. */
  sectionCount?: number;
}

export function CourseCard({
  course,
  variant,
  href,
  rightSlot,
  memberCount,
  sectionCount,
}: CourseCardProps) {
  const destination = href ?? `/courses/${course.id}`;
  const code = `${course.department} ${course.number}`;

  return (
    <div
      className={cn(TILE_SHELL_CLASSES, "hover:shadow-md focus-within:ring-2")}
    >
      <Link
        href={destination}
        aria-label={`${code} — ${course.title}`}
        className="absolute inset-0 z-0 rounded-[10px] focus:outline-none"
      />
      {variant === "row" ? (
        <RowVariant course={course} code={code} rightSlot={rightSlot} />
      ) : (
        <TileVariant
          course={course}
          code={code}
          rightSlot={rightSlot}
          memberCount={memberCount}
          sectionCount={sectionCount}
        />
      )}
    </div>
  );
}

/**
 * Loading placeholder that mirrors {@link CourseCard}'s tile shape.
 * Lives next to the real card so the two share `TILE_SHELL_CLASSES`
 * + `TILE_INNER_CLASSES` and can't drift apart silently.
 */
export function CourseCardSkeleton() {
  return (
    <div className={TILE_SHELL_CLASSES} aria-hidden={true}>
      <div className={cn(TILE_INNER_CLASSES, "pointer-events-none")}>
        <div className="flex items-start justify-between gap-2">
          <Skeleton className="h-3.5 w-16" />
        </div>
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/5" />
        <Skeleton className="h-3.5 w-2/5" />
        <div className="flex-1" />
        <div className="flex justify-end">
          <Skeleton className="size-3.5 rounded-full" />
        </div>
      </div>
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
    <div className="pointer-events-none flex items-center gap-3 p-3">
      <div className="relative z-10 min-w-0 flex-1">
        <p className="text-foreground text-sm font-semibold">{code}</p>
        <p
          className="text-muted-foreground truncate text-xs"
          title={course.title}
        >
          {course.title}
        </p>
      </div>
      {rightSlot ? (
        <div className="pointer-events-auto relative z-10 shrink-0">
          {rightSlot}
        </div>
      ) : null}
    </div>
  );
}

function TileVariant({
  course,
  code,
  rightSlot,
  memberCount,
  sectionCount,
}: {
  course: CourseResponse;
  code: string;
  rightSlot?: ReactNode;
  memberCount?: number;
  sectionCount?: number;
}) {
  const hasFooterMetrics =
    typeof memberCount === "number" || typeof sectionCount === "number";

  return (
    <div className={cn(TILE_INNER_CLASSES, "pointer-events-none")}>
      <div className="flex items-start justify-between gap-2">
        <span className="text-foreground font-mono text-[13px] font-semibold tracking-[-0.2px]">
          {code}
        </span>
        {rightSlot ? (
          <div className="pointer-events-auto relative z-10 shrink-0">
            {rightSlot}
          </div>
        ) : null}
      </div>

      <h3 className="text-foreground line-clamp-2 text-[15px] font-semibold leading-[1.35] tracking-[-0.2px]">
        {course.title}
      </h3>

      <p className="truncate text-[13px] text-zinc-500">{course.school.name}</p>

      <div className="flex-1" />

      {hasFooterMetrics ? (
        <>
          <div className="h-px w-full bg-zinc-100" />
          <div className="flex items-center justify-between pt-1.5">
            <div className="flex items-center gap-3.5 text-[12px] font-medium text-zinc-600">
              {typeof memberCount === "number" ? (
                <span className="flex items-center gap-1.5">
                  <Users className="size-3 text-zinc-400" aria-hidden={true} />
                  {memberCount}
                </span>
              ) : null}
              {typeof sectionCount === "number" ? (
                <span className="flex items-center gap-1.5">
                  <Layers className="size-3 text-zinc-400" aria-hidden={true} />
                  {sectionCount} {sectionCount === 1 ? "section" : "sections"}
                </span>
              ) : null}
            </div>
            <ArrowRight
              className="size-3.5 text-zinc-400 transition-transform group-hover:translate-x-0.5"
              aria-hidden={true}
            />
          </div>
        </>
      ) : (
        <div className="flex justify-end">
          <ArrowRight
            className="size-3.5 text-zinc-400 transition-transform group-hover:translate-x-0.5"
            aria-hidden={true}
          />
        </div>
      )}
    </div>
  );
}
