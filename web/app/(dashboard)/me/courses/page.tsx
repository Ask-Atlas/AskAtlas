/**
 * "My courses" — landing for the user's enrolled sections.
 *
 * Server Component. Fetches `listMyEnrollments` once and groups the
 * rows by term so the user sees their current-term load at the top
 * with previous terms below. Each card opens the canonical course
 * detail page; section + role surface inline so the user knows which
 * section they're in (a course can have multiple sections, and the
 * adopted-demo user is enrolled in all of them).
 */
import { CalendarOff, GraduationCap, UserRound } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { listMyEnrollments } from "@/lib/api";
import type { EnrollmentResponse } from "@/lib/api/types";
import { cn } from "@/lib/utils";

export default async function MyCoursesPage() {
  const { enrollments } = await listMyEnrollments();
  const groups = groupByTerm(enrollments);

  return (
    <section className="flex flex-col gap-8 py-2">
      <header className="space-y-1.5">
        <div className="flex items-center gap-3">
          <h1 className="text-foreground text-[28px] font-semibold leading-tight tracking-[-0.4px]">
            My courses
          </h1>
          <span className="bg-muted text-muted-foreground rounded-md px-2 py-0.5 font-mono text-[12px] font-semibold">
            {enrollments.length}
          </span>
        </div>
        <p className="text-muted-foreground text-sm">
          Sections you&rsquo;re currently in, grouped by term.
        </p>
      </header>

      {groups.length === 0 ? (
        <EmptyState
          icon={<CalendarOff className="size-8" aria-hidden={true} />}
          title="You&rsquo;re not enrolled in any sections yet"
          body="Browse the catalog and join a section to see study guides, files, and quizzes for that course."
          action={
            <Button asChild>
              <Link href="/courses">Browse courses</Link>
            </Button>
          }
          className="border-border bg-muted/30 rounded-[10px] border py-14"
        />
      ) : (
        <div className="flex flex-col gap-10">
          {groups.map((group) => (
            <TermBlock key={group.term} group={group} />
          ))}
        </div>
      )}
    </section>
  );
}

interface TermGroup {
  term: string;
  enrollments: EnrollmentResponse[];
}

function groupByTerm(enrollments: EnrollmentResponse[]): TermGroup[] {
  const byTerm = new Map<string, EnrollmentResponse[]>();
  for (const e of enrollments) {
    const list = byTerm.get(e.section.term) ?? [];
    list.push(e);
    byTerm.set(e.section.term, list);
  }
  return Array.from(byTerm.entries())
    .map(([term, list]) => ({ term, enrollments: list }))
    .sort((a, b) => b.term.localeCompare(a.term));
}

function TermBlock({ group }: { group: TermGroup }) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <h2 className="text-foreground text-[18px] font-semibold leading-tight tracking-[-0.3px]">
          {group.term}
        </h2>
        <p className="text-muted-foreground text-[12px]">
          {group.enrollments.length}{" "}
          {group.enrollments.length === 1 ? "section" : "sections"}
        </p>
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {group.enrollments.map((e) => (
          <EnrollmentCard
            key={`${e.course.id}-${e.section.id}`}
            enrollment={e}
          />
        ))}
      </div>
    </div>
  );
}

function EnrollmentCard({ enrollment }: { enrollment: EnrollmentResponse }) {
  const { course, section, role, school } = enrollment;
  const sectionLabel = section.section_code
    ? `Section ${section.section_code}`
    : "Your section";
  const roleLabel = roleDisplay(role);
  return (
    <Link
      href={`/courses/${course.id}`}
      className={cn(
        "bg-card border-border group flex h-full flex-col gap-3 rounded-[10px] border p-4 transition-all",
        "hover:border-foreground/20 focus-visible:ring-ring hover:shadow-md focus-visible:outline-hidden focus-visible:ring-2",
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="text-foreground bg-muted/60 inline-flex items-center rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold tracking-[-0.2px]">
          {course.department} {course.number}
        </div>
        <span
          className={cn(
            "rounded-full px-2 py-0.5 text-[11px] font-medium",
            role === "instructor" || role === "ta"
              ? "bg-primary/10 text-primary"
              : "bg-muted text-muted-foreground",
          )}
        >
          {roleLabel}
        </span>
      </div>
      <h3 className="text-foreground line-clamp-2 text-[15px] font-semibold leading-snug tracking-[-0.2px]">
        {course.title}
      </h3>
      <div className="text-muted-foreground mt-auto flex flex-col gap-1.5 text-[12px]">
        <span className="inline-flex items-center gap-1.5">
          <GraduationCap
            className="text-muted-foreground/70 size-3.5"
            aria-hidden={true}
          />
          {school.acronym} · {sectionLabel}
        </span>
        {section.instructor_name ? (
          <span className="inline-flex items-center gap-1.5">
            <UserRound
              className="text-muted-foreground/70 size-3.5"
              aria-hidden={true}
            />
            {section.instructor_name}
          </span>
        ) : null}
      </div>
    </Link>
  );
}

function roleDisplay(role: EnrollmentResponse["role"]): string {
  if (role === "instructor") return "Instructor";
  if (role === "ta") return "TA";
  return "Student";
}
