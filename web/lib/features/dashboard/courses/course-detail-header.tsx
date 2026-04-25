import { Calendar, FileText, Layers, Share2, Star, Users } from "lucide-react";
import { type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import type { CourseDetailResponse } from "@/lib/api/types";

interface CourseDetailHeaderProps {
  course: CourseDetailResponse;
  /** Total enrolled across sections — `sum(section.member_count)`. */
  enrolledCount: number;
  studyGuidesCount: number;
  /** Optional term label rendered next to the dept code (e.g. "Spring 2026"). */
  termLabel?: string;
  /** Optional formatted "Last updated" string. */
  lastUpdatedLabel?: string;
  /** Optional trailing action(s) placed next to Star + Share. */
  trailingActions?: ReactNode;
}

export function CourseDetailHeader({
  course,
  enrolledCount,
  studyGuidesCount,
  termLabel,
  lastUpdatedLabel,
  trailingActions,
}: CourseDetailHeaderProps) {
  const code = `${course.department} ${course.number}`;
  const sectionsCount = course.sections.length;

  return (
    <header className="flex flex-col gap-3.5">
      <div className="flex items-start justify-between gap-6">
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          <div className="flex items-center gap-2.5">
            <span className="text-foreground font-mono text-[13px] font-semibold tracking-[-0.2px]">
              {code}
            </span>
            {termLabel ? (
              <>
                <span
                  className="bg-border size-[3px] rounded-full"
                  aria-hidden={true}
                />
                <span className="text-muted-foreground text-[13px]">
                  {termLabel}
                </span>
              </>
            ) : null}
          </div>
          <h1 className="text-foreground text-[30px] font-semibold leading-[1.15] tracking-[-0.6px]">
            {course.title}
          </h1>
          <p className="text-muted-foreground text-[14px]">
            {course.school.name}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="icon"
            aria-label="Favorite"
            className="size-9"
          >
            <Star className="size-4" aria-hidden={true} />
          </Button>
          <Button
            type="button"
            variant="outline"
            size="icon"
            aria-label="Share"
            className="size-9"
          >
            <Share2 className="size-4" aria-hidden={true} />
          </Button>
          {trailingActions}
        </div>
      </div>

      <div className="bg-border h-px w-full" />

      <div className="text-muted-foreground flex flex-wrap items-center gap-x-[18px] gap-y-2 text-[13px] font-medium">
        <span className="flex items-center gap-1.5">
          <Users
            className="text-muted-foreground/60 size-3.5"
            aria-hidden={true}
          />
          {enrolledCount.toLocaleString()} enrolled
        </span>
        <span className="flex items-center gap-1.5">
          <Layers
            className="text-muted-foreground/60 size-3.5"
            aria-hidden={true}
          />
          {sectionsCount} {sectionsCount === 1 ? "section" : "sections"}
        </span>
        <span className="flex items-center gap-1.5">
          <FileText
            className="text-muted-foreground/60 size-3.5"
            aria-hidden={true}
          />
          {studyGuidesCount}{" "}
          {studyGuidesCount === 1 ? "study guide" : "study guides"}
        </span>
        {lastUpdatedLabel ? (
          <span className="flex items-center gap-1.5">
            <Calendar
              className="text-muted-foreground/60 size-3.5"
              aria-hidden={true}
            />
            {lastUpdatedLabel}
          </span>
        ) : null}
      </div>
    </header>
  );
}
