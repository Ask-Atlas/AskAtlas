import { BookOpen, Plus } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { listMyStudyGuides } from "@/lib/api";
import type { ListMyStudyGuidesResponse } from "@/lib/api/types";
import {
  StudyGuideCard,
  type StudyGuideListItemResponse as StudyGuideCardListItem,
} from "@/lib/features/dashboard/study-guides/study-guide-card";

type MyStudyGuide = ListMyStudyGuidesResponse["study_guides"][number];

export default async function MyStudyGuidesPage() {
  const res = await listMyStudyGuides();
  // The endpoint includes soft-deleted guides for future restore UX (ASK-131);
  // until that lands, hide them from the live list.
  const guides = res.study_guides.filter((g) => g.deleted_at === null);

  return (
    <section className="flex flex-col gap-8 py-2">
      <header className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
        <div className="space-y-1.5">
          <div className="flex items-center gap-3">
            <h1 className="text-foreground text-[28px] font-semibold leading-tight tracking-[-0.4px]">
              My study guides
            </h1>
            <span className="bg-muted text-muted-foreground rounded-md px-2 py-0.5 font-mono text-[12px] font-semibold">
              {guides.length}
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Drafts and guides you&rsquo;ve written. Looking for everything in
            your courses?{" "}
            <Link
              href="/study-guides"
              className="text-foreground font-medium underline-offset-2 hover:underline"
            >
              Browse by course
            </Link>
            .
          </p>
        </div>
        <Button asChild className="h-9">
          <Link href="/study-guides/new">
            <Plus className="size-4" aria-hidden={true} />
            New guide
          </Link>
        </Button>
      </header>

      {guides.length === 0 ? (
        <EmptyState
          icon={<BookOpen className="size-8" aria-hidden={true} />}
          title="You haven&rsquo;t written any guides yet"
          body="Start your first one, or browse what's already in your courses."
          action={
            <div className="flex flex-wrap items-center justify-center gap-2">
              <Button asChild>
                <Link href="/study-guides/new">Create one</Link>
              </Button>
              <Button asChild variant="outline">
                <Link href="/study-guides">Browse by course</Link>
              </Button>
            </div>
          }
          className="border-border bg-muted/30 rounded-[10px] border py-14"
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {guides.map((guide) => (
            <StudyGuideCard
              key={guide.id}
              guide={toCardItem(guide)}
              variant="list"
            />
          ))}
        </div>
      )}
    </section>
  );
}

function toCardItem(guide: MyStudyGuide): StudyGuideCardListItem {
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
