import Link from "next/link";
import { type KeyboardEvent, type MouseEvent, type ReactNode } from "react";

import type { CourseResponse } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface CourseCardProps {
  course: CourseResponse;
  variant: "row" | "tile";
  /** Defaults to `/courses/{id}`. Override for contextual routes (e.g. a catalog filter). */
  href?: string;
  /**
   * Trailing affordance: "Joined" badge, FavoriteButton, enrollment CTA, etc.
   * Clicks and keypresses inside this slot do NOT navigate so interactive
   * children can own their own behavior without the outer Link swallowing them.
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

  return (
    <Link
      href={destination}
      className={cn(
        "group bg-card block rounded-xl border transition-all",
        "hover:shadow-md",
        "focus-visible:ring-ring focus-visible:outline-none focus-visible:ring-2",
        variant === "row" ? "p-3" : "p-4",
      )}
    >
      {variant === "row" ? (
        <RowVariant course={course} rightSlot={rightSlot} />
      ) : (
        <TileVariant course={course} rightSlot={rightSlot} />
      )}
    </Link>
  );
}

function RowVariant({
  course,
  rightSlot,
}: {
  course: CourseResponse;
  rightSlot?: ReactNode;
}) {
  const code = `${course.department} ${course.number}`;
  return (
    <div className="flex items-center gap-3">
      <div className="min-w-0 flex-1">
        <p className="text-foreground text-sm font-semibold">{code}</p>
        <p
          className="text-muted-foreground truncate text-xs"
          title={course.title}
        >
          {course.title}
        </p>
      </div>
      {rightSlot ? <SlotWrapper>{rightSlot}</SlotWrapper> : null}
    </div>
  );
}

function TileVariant({
  course,
  rightSlot,
}: {
  course: CourseResponse;
  rightSlot?: ReactNode;
}) {
  const code = `${course.department} ${course.number}`;
  return (
    <div className="relative space-y-1.5">
      {rightSlot ? (
        <SlotWrapper className="absolute right-0 top-0">
          {rightSlot}
        </SlotWrapper>
      ) : null}
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
  );
}

function SlotWrapper({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn("shrink-0", className)}
      onClick={stopPropagation}
      onKeyDown={stopKeyboardPropagation}
    >
      {children}
    </div>
  );
}

function stopPropagation(event: MouseEvent) {
  event.stopPropagation();
}

function stopKeyboardPropagation(event: KeyboardEvent) {
  event.stopPropagation();
}
