/**
 * Course detail page (ASK-197).
 *
 * Server Component. Fetches course detail, sections, study guides, and
 * the caller's enrollments in parallel; the enrollment list lets us
 * pre-resolve per-section membership without an N+1 of `checkMembership`
 * round-trips. A 404 from `getCourse` triggers the dashboard not-found
 * boundary; any other failure bubbles to the dashboard error boundary.
 *
 * The page renders one of two states based on whether the caller is in a
 * section of this course:
 *   - Enrolled: section status banner + study-guide grid (the body).
 *   - Not enrolled: "Pick a section" picker + dimmed study-guide teaser.
 */
import {
  ArrowRightLeft,
  BookOpen,
  CalendarOff,
  Lock,
  Plus,
  UserRound,
  Users,
} from "lucide-react";
import Link from "next/link";
import { notFound } from "next/navigation";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { getCourse, listCourseStudyGuides, listMyEnrollments } from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import type {
  SectionSummary,
  StudyGuideListItemResponse,
} from "@/lib/api/types";
import { CourseDetailHeader } from "@/lib/features/dashboard/courses/course-detail-header";
import { SectionRow } from "@/lib/features/dashboard/courses/section-row";
import {
  StudyGuideCard,
  type StudyGuideListItemResponse as StudyGuideCardListItem,
} from "@/lib/features/dashboard/study-guides/study-guide-card";

interface PageProps {
  params: Promise<{ courseId: string }>;
}

export default async function CourseDetailPage({ params }: PageProps) {
  const { courseId } = await params;

  // Sections come embedded on `CourseDetailResponse`, so we don't also
  // hit `listCourseSections` -- the dedicated endpoint adds `course_id`
  // and `created_at` that this page doesn't render. Enrollments power
  // per-section membership without an N+1 of `checkMembership` calls.
  const [course, studyGuidesRes, enrollments] = await Promise.all([
    getCourseOr404(courseId),
    listCourseStudyGuides(courseId),
    listMyEnrollments(),
  ]);

  const sectionIds = new Set(course.sections.map((s) => s.id));
  const myEnrollmentInCourse =
    enrollments.enrollments.find((e) => sectionIds.has(e.section.id)) ?? null;

  const enrolledCount = course.sections.reduce(
    (acc, s) => acc + s.member_count,
    0,
  );
  const studyGuides = studyGuidesRes.study_guides;
  const isEnrolled = Boolean(myEnrollmentInCourse);

  return (
    <section className="mx-auto flex w-full max-w-[1184px] flex-col gap-7 px-10 py-8">
      <CourseDetailHeader
        course={course}
        enrolledCount={enrolledCount}
        studyGuidesCount={studyGuides.length}
        termLabel={myEnrollmentInCourse?.section.term}
        trailingActions={
          isEnrolled && myEnrollmentInCourse ? (
            <Button asChild variant="outline" className="h-9">
              <Link
                href={`/courses/${course.id}/sections/${myEnrollmentInCourse.section.id}/members`}
              >
                <Users className="size-4" aria-hidden={true} />
                Members
              </Link>
            </Button>
          ) : null
        }
      />

      {isEnrolled && myEnrollmentInCourse ? (
        <>
          <EnrolledBanner
            courseId={course.id}
            sectionId={myEnrollmentInCourse.section.id}
            sectionCode={myEnrollmentInCourse.section.section_code}
            instructor={myEnrollmentInCourse.section.instructor_name}
            term={myEnrollmentInCourse.section.term}
          />
          <StudyGuidesSection courseId={course.id} studyGuides={studyGuides} />
        </>
      ) : (
        <>
          <SectionPicker courseId={course.id} sections={course.sections} />
          <StudyGuidesTeaser studyGuides={studyGuides} />
        </>
      )}
    </section>
  );
}

async function getCourseOr404(courseId: string) {
  try {
    return await getCourse(courseId);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      notFound();
    }
    throw err;
  }
}

function EnrolledBanner({
  courseId,
  sectionId,
  sectionCode,
  instructor,
  term,
}: {
  courseId: string;
  sectionId: string;
  sectionCode: string | null | undefined;
  instructor: string | null | undefined;
  term: string;
}) {
  const sectionLabel = sectionCode ? `Section ${sectionCode}` : "your section";
  return (
    <div className="border-border bg-muted/40 flex flex-wrap items-center justify-between gap-4 rounded-[10px] border px-3 py-3">
      <div className="flex min-w-0 items-center gap-3.5">
        <span
          className="bg-primary/10 text-primary flex size-8 shrink-0 items-center justify-center rounded-lg"
          aria-hidden={true}
        >
          <UserRound className="size-4" />
        </span>
        <div className="flex min-w-0 flex-col gap-0.5">
          <p className="flex flex-wrap items-center gap-x-1.5 text-[13px]">
            <span className="text-muted-foreground">You&rsquo;re in</span>
            <span className="text-foreground font-mono font-semibold">
              {sectionLabel}
            </span>
            {instructor ? (
              <>
                <span className="text-muted-foreground/50" aria-hidden={true}>
                  ·
                </span>
                <span className="text-foreground font-medium">
                  {instructor}
                </span>
              </>
            ) : null}
          </p>
          <p className="text-muted-foreground text-[12px]">{term}</p>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <Button asChild variant="ghost" size="sm" className="h-8">
          <Link href={`/courses/${courseId}/sections/${sectionId}/members`}>
            <Users className="size-3.5" aria-hidden={true} />
            Section roster
          </Link>
        </Button>
        <Button variant="ghost" size="sm" className="h-8">
          <ArrowRightLeft className="size-3.5" aria-hidden={true} />
          Switch
        </Button>
      </div>
    </div>
  );
}

function SectionPicker({
  courseId,
  sections,
}: {
  courseId: string;
  sections: SectionSummary[];
}) {
  if (sections.length === 0) {
    return (
      <EmptyState
        icon={<CalendarOff className="size-8" aria-hidden={true} />}
        title="No sections offered"
        body="This course has no current section offerings. Check back next term."
        className="border-border bg-muted/30 rounded-[10px] border py-12"
      />
    );
  }
  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-col gap-1.5">
        <h2 className="text-foreground text-[18px] font-semibold leading-tight tracking-[-0.3px]">
          Pick a section
        </h2>
        <p className="text-muted-foreground text-[13px]">
          Join a section to access study guides for this course. You&rsquo;ll
          only see what your section is sharing.
        </p>
      </div>
      <div className="divide-border bg-card border-border flex flex-col divide-y rounded-[10px] border">
        {sections.map((section) => (
          <SectionRow
            key={section.id}
            courseId={courseId}
            section={section}
            initialMembership="not-member"
          />
        ))}
      </div>
    </div>
  );
}

function StudyGuidesSection({
  courseId,
  studyGuides,
}: {
  courseId: string;
  studyGuides: StudyGuideListItemResponse[];
}) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-3">
            <h2 className="text-foreground text-[22px] font-semibold leading-tight tracking-[-0.4px]">
              Study guides
            </h2>
            <span className="bg-muted text-muted-foreground rounded-md px-2 py-0.5 font-mono text-[12px] font-semibold">
              {studyGuides.length}
            </span>
          </div>
          <p className="text-muted-foreground text-[13px]">
            Browse what your classmates have written, or start your own.
          </p>
        </div>
        <Button asChild className="h-9">
          <Link href={`/study-guides/new?course=${courseId}`}>
            <Plus className="size-4" aria-hidden={true} />
            New guide
          </Link>
        </Button>
      </div>

      {studyGuides.length === 0 ? (
        <EmptyState
          icon={<BookOpen className="size-8" aria-hidden={true} />}
          title="No study guides yet"
          body="Be the first to share notes, an outline, or a cheat sheet with your section."
          action={
            <Button asChild>
              <Link href={`/study-guides/new?course=${courseId}`}>
                Create one
              </Link>
            </Button>
          }
          className="border-border bg-muted/30 rounded-[10px] border py-14"
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {studyGuides.map((guide) => (
            <StudyGuideCard
              key={guide.id}
              guide={toCardItem(guide)}
              variant="list"
            />
          ))}
        </div>
      )}
    </div>
  );
}

function StudyGuidesTeaser({
  studyGuides,
}: {
  studyGuides: StudyGuideListItemResponse[];
}) {
  if (studyGuides.length === 0) {
    return null;
  }
  const previews = studyGuides.slice(0, 3);
  const lockedCount = Math.max(studyGuides.length - previews.length, 0);
  const opacities = ["opacity-60", "opacity-40", "opacity-25"];
  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center gap-3">
          <h2 className="text-muted-foreground text-[22px] font-semibold leading-tight tracking-[-0.4px]">
            Study guides
          </h2>
          <span className="bg-muted text-muted-foreground rounded-md px-2 py-0.5 font-mono text-[12px] font-semibold">
            {studyGuides.length}
          </span>
        </div>
        <p className="text-muted-foreground/80 text-[13px]">
          You&rsquo;ll see what your section is sharing once you join.
        </p>
      </div>
      <div
        className="pointer-events-none grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
        aria-hidden={true}
      >
        {previews.map((guide, index) => (
          <div key={guide.id} className={opacities[index] ?? "opacity-25"}>
            <StudyGuideCard guide={toCardItem(guide)} variant="list" />
          </div>
        ))}
      </div>
      {lockedCount > 0 ? (
        <p className="text-muted-foreground flex items-center justify-center gap-1.5 pt-1 text-[13px] font-medium">
          <Lock
            className="text-muted-foreground/60 size-3.5"
            aria-hidden={true}
          />
          + {lockedCount} more {lockedCount === 1 ? "guide" : "guides"} locked
        </p>
      ) : null}
    </div>
  );
}

function toCardItem(guide: StudyGuideListItemResponse): StudyGuideCardListItem {
  const displayName =
    `${guide.creator.first_name} ${guide.creator.last_name}`.trim() ||
    "Unknown author";
  return {
    id: guide.id,
    title: guide.title,
    creator: { display_name: displayName },
    vote_score: guide.vote_score,
    quiz_count: guide.quiz_count,
    is_recommended: guide.is_recommended,
    tags: guide.tags,
    course_id: guide.course_id,
    visibility: guide.visibility,
  };
}
