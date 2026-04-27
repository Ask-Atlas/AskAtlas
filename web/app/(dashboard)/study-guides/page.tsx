/**
 * Browse-by-course study guide landing page.
 *
 * Server Component. Fetches the caller's enrollments + every enrolled
 * course's guide list in parallel, then renders one section per
 * course with a StudyGuideCard grid. Empty courses surface as small
 * "no guides yet" placeholders so the user still gets a per-course
 * affordance to author one.
 */
import { BookOpen, Plus } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { listCourseStudyGuides, listMyEnrollments } from "@/lib/api";
import type {
  ListStudyGuidesResponse,
  StudyGuideListItemResponse,
} from "@/lib/api/types";
import {
  StudyGuideCard,
  type StudyGuideListItemResponse as StudyGuideCardListItem,
} from "@/lib/features/dashboard/study-guides/study-guide-card";

interface CourseGuides {
  courseId: string;
  courseLabel: string;
  courseTitle: string;
  guides: StudyGuideListItemResponse[];
}

export default async function StudyGuidesPage() {
  const { enrollments } = await listMyEnrollments();
  // Dedupe by course id — a user enrolled in two sections of the
  // same course should still see one course block.
  const uniqueCourses = new Map<string, { label: string; title: string }>();
  for (const e of enrollments) {
    if (!uniqueCourses.has(e.course.id)) {
      uniqueCourses.set(e.course.id, {
        label: `${e.course.department} ${e.course.number}`,
        title: e.course.title,
      });
    }
  }

  const courseEntries = Array.from(uniqueCourses.entries());
  const guideResponses = await Promise.all(
    courseEntries.map(([courseId]) => listCourseStudyGuides(courseId)),
  );

  const blocks: CourseGuides[] = courseEntries.map(
    ([courseId, meta], index): CourseGuides => ({
      courseId,
      courseLabel: meta.label,
      courseTitle: meta.title,
      guides: (guideResponses[index] as ListStudyGuidesResponse).study_guides,
    }),
  );

  const totalGuides = blocks.reduce((sum, b) => sum + b.guides.length, 0);

  return (
    <section className="flex flex-col gap-8 py-2">
      <header className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
        <div className="space-y-1.5">
          <div className="flex items-center gap-3">
            <h1 className="text-foreground text-[28px] font-semibold leading-tight tracking-[-0.4px]">
              Study guides
            </h1>
            <span className="bg-muted text-muted-foreground rounded-md px-2 py-0.5 font-mono text-[12px] font-semibold">
              {totalGuides}
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Browse guides across the courses you&rsquo;re enrolled in.
          </p>
        </div>
        <Button asChild className="h-9">
          <Link href="/study-guides/new">
            <Plus className="size-4" aria-hidden={true} />
            New guide
          </Link>
        </Button>
      </header>

      {blocks.length === 0 ? (
        <EmptyState
          icon={<BookOpen className="size-8" aria-hidden={true} />}
          title="You&rsquo;re not enrolled in any courses yet"
          body="Browse the catalog and join a section to see study guides here."
          action={
            <Button asChild>
              <Link href="/courses">Browse courses</Link>
            </Button>
          }
          className="border-border bg-muted/30 rounded-[10px] border py-14"
        />
      ) : (
        <div className="flex flex-col gap-10">
          {blocks.map((block) => (
            <CourseBlock key={block.courseId} block={block} />
          ))}
        </div>
      )}
    </section>
  );
}

function CourseBlock({ block }: { block: CourseGuides }) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="flex flex-col gap-0.5">
          <Link
            href={`/courses/${block.courseId}`}
            className="text-foreground hover:underline"
          >
            <span className="text-muted-foreground bg-muted/60 inline-flex items-center rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold tracking-[-0.2px]">
              {block.courseLabel}
            </span>
            <h2 className="text-foreground mt-1 text-[18px] font-semibold leading-tight tracking-[-0.3px]">
              {block.courseTitle}
            </h2>
          </Link>
        </div>
        <p className="text-muted-foreground text-[12px]">
          {block.guides.length} {block.guides.length === 1 ? "guide" : "guides"}
        </p>
      </div>
      {block.guides.length === 0 ? (
        <p className="text-muted-foreground border-border bg-muted/30 rounded-[10px] border px-4 py-5 text-center text-sm">
          No guides in this course yet.{" "}
          <Link
            href={`/study-guides/new?course=${block.courseId}`}
            className="text-foreground font-medium underline-offset-2 hover:underline"
          >
            Be the first.
          </Link>
        </p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {block.guides.map((guide) => (
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

function toCardItem(guide: StudyGuideListItemResponse): StudyGuideCardListItem {
  const displayName =
    `${guide.creator.first_name} ${guide.creator.last_name}`.trim() ||
    "Unknown author";
  return {
    id: guide.id,
    title: guide.title,
    description: guide.description,
    creator: { display_name: displayName },
    vote_score: guide.vote_score,
    quiz_count: guide.quiz_count,
    is_recommended: guide.is_recommended,
    tags: guide.tags,
    course_id: guide.course_id,
    visibility: guide.visibility,
    updated_at: guide.updated_at,
  };
}
