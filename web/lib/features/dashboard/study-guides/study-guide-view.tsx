"use client";

/**
 * Reader surface for a single study guide. Renders the guide header,
 * the article body via `<ArticleRenderer>` (markdown + GFM + embedded
 * images via `/api/files/{id}/download`), and inline lists for
 * quizzes, resources, and attached files. Author-only affordances
 * (Edit, Delete) gate on the `canEdit` prop the page resolves from
 * Clerk.
 */
import {
  ExternalLink,
  FileText,
  Pencil,
  Play,
  ThumbsUp,
  Trash2,
  Video,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Separator } from "@/components/ui/separator";
import { deleteStudyGuide, recordFileView } from "@/lib/api";
import type {
  ResourceSummary,
  StudyGuideDetailResponse,
  StudyGuideFileSummary,
} from "@/lib/api/types";
import { ConfirmationDialog } from "@/lib/features/shared/confirmation-dialog";
import { toast } from "@/lib/features/shared/toast/toast";
import { cn, formatBytes, formatRelativeDate } from "@/lib/utils";

import { ArticleRenderer } from "./article-renderer";

interface StudyGuideViewProps {
  guide: StudyGuideDetailResponse;
  canEdit: boolean;
}

export function StudyGuideView({ guide, canEdit }: StudyGuideViewProps) {
  const router = useRouter();
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [isDeleting, startDeleteTransition] = useTransition();

  const creatorName =
    `${guide.creator.first_name} ${guide.creator.last_name}`.trim() ||
    "Unknown author";
  const courseLabel = `${guide.course.department} ${guide.course.number}`;

  const handleConfirmDelete = () => {
    startDeleteTransition(async () => {
      try {
        await deleteStudyGuide(guide.id);
        toast.success("Study guide deleted");
        router.push(`/courses/${guide.course.id}`);
      } catch (err) {
        toast.error(err);
      } finally {
        setConfirmDeleteOpen(false);
      }
    });
  };

  return (
    <section className="mx-auto flex w-full max-w-3xl flex-col gap-8 py-2">
      <header className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <Link
            href={`/courses/${guide.course.id}`}
            className="text-foreground hover:bg-muted bg-muted/60 inline-flex items-center rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold tracking-[-0.2px] transition-colors"
          >
            {courseLabel}
          </Link>
          <Badge variant="secondary" className="text-[11px]">
            {guide.visibility === "public" ? "Public" : "Private"}
          </Badge>
        </div>
        <h1 className="text-foreground text-[32px] font-semibold leading-[1.15] tracking-[-0.6px]">
          {guide.title}
        </h1>
        {guide.description ? (
          <p className="text-muted-foreground text-[15px] leading-[1.55]">
            {guide.description}
          </p>
        ) : null}
        <p className="text-muted-foreground flex flex-wrap items-center gap-x-2 text-[13px]">
          <span>by {creatorName}</span>
          <span aria-hidden={true} className="text-muted-foreground/50">
            ·
          </span>
          <span>Updated {formatRelativeDate(guide.updated_at)}</span>
          <span aria-hidden={true} className="text-muted-foreground/50">
            ·
          </span>
          <span className="inline-flex items-center gap-1">
            <ThumbsUp className="size-3.5" aria-hidden={true} />
            {guide.vote_score}
          </span>
        </p>
        {canEdit ? (
          <div className="flex flex-wrap gap-2 pt-2">
            <Button asChild variant="outline" size="sm" className="h-8">
              <Link href={`/study-guides/${guide.id}/edit`}>
                <Pencil className="size-3.5" aria-hidden={true} />
                Edit
              </Link>
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive h-8"
              onClick={() => setConfirmDeleteOpen(true)}
              disabled={isDeleting}
            >
              <Trash2 className="size-3.5" aria-hidden={true} />
              Delete
            </Button>
          </div>
        ) : null}
      </header>

      {guide.content ? (
        <ArticleRenderer content={guide.content} />
      ) : (
        <EmptyState
          icon={<FileText className="size-7" aria-hidden={true} />}
          title="This guide is empty"
          body={
            canEdit
              ? "Add some content via the editor."
              : "The author hasn't written anything yet."
          }
          className="border-border bg-muted/30 rounded-[10px] border py-10"
        />
      )}

      {guide.tags.length > 0 ? (
        <div className="flex flex-wrap gap-1.5">
          {guide.tags.map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="bg-muted text-muted-foreground text-[11px]"
            >
              {tag}
            </Badge>
          ))}
        </div>
      ) : null}

      <Section title="Attached files" count={guide.files.length}>
        {guide.files.length === 0 ? (
          <SectionEmpty body="No files attached yet." />
        ) : (
          <div className="border-border divide-border overflow-hidden divide-y rounded-[10px] border">
            {guide.files.map((file) => (
              <FileRow key={file.id} file={file} />
            ))}
          </div>
        )}
      </Section>

      <Section title="Resources" count={guide.resources.length}>
        {guide.resources.length === 0 ? (
          <SectionEmpty body="No resources linked yet." />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {guide.resources.map((resource) => (
              <ResourceRow key={resource.id} resource={resource} />
            ))}
          </div>
        )}
      </Section>

      <Section title="Quizzes" count={guide.quizzes.length}>
        {guide.quizzes.length === 0 ? (
          <SectionEmpty body="No quizzes yet." />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {guide.quizzes.map((quiz) => (
              <div
                key={quiz.id}
                className="bg-card border-border flex items-center justify-between gap-3 rounded-[10px] border p-3"
              >
                <div className="min-w-0 flex-1">
                  <p className="text-foreground truncate text-sm font-semibold">
                    {quiz.title}
                  </p>
                  <p className="text-muted-foreground text-xs">
                    {quiz.question_count}{" "}
                    {quiz.question_count === 1 ? "question" : "questions"}
                  </p>
                </div>
                <Button asChild size="sm" className="h-8 shrink-0">
                  <Link href={`/practice?quiz=${quiz.id}`}>
                    <Play className="size-3" aria-hidden={true} />
                    Start
                  </Link>
                </Button>
              </div>
            ))}
          </div>
        )}
      </Section>

      <ConfirmationDialog
        open={confirmDeleteOpen}
        onOpenChange={setConfirmDeleteOpen}
        title="Delete this study guide?"
        description="This permanently deletes the guide, its quizzes, and all attached files / resources for everyone who can see it. This can't be undone."
        confirmLabel={isDeleting ? "Deleting…" : "Delete"}
        cancelLabel="Cancel"
        destructive
        disabled={isDeleting}
        onConfirm={handleConfirmDelete}
      />
    </section>
  );
}

function Section({
  title,
  count,
  children,
}: {
  title: string;
  count: number;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2.5">
        <h2 className="text-foreground text-[18px] font-semibold leading-tight tracking-[-0.3px]">
          {title}
        </h2>
        <span className="bg-muted text-muted-foreground rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold">
          {count}
        </span>
      </div>
      <Separator className="opacity-60" />
      {children}
    </div>
  );
}

function SectionEmpty({ body }: { body: string }) {
  return (
    <div className="border-border bg-muted/30 text-muted-foreground rounded-[10px] border px-4 py-6 text-center text-sm">
      {body}
    </div>
  );
}

function FileRow({ file }: { file: StudyGuideFileSummary }) {
  const handleOpen = () => {
    // recordFileView is fire-and-forget so a slow recents update never
    // blocks the user from reading the file.
    void recordFileView(file.id).catch(() => {
      // intentional: telemetry failure must not surface to the user.
    });
    window.open(`/api/files/${file.id}/download`, "_blank", "noopener");
  };
  return (
    <button
      type="button"
      onClick={handleOpen}
      className={cn(
        "hover:bg-muted/40 flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors",
        "focus-visible:bg-muted/40 focus-visible:outline-none",
      )}
    >
      <div className="flex min-w-0 items-center gap-3">
        <FileText
          className="text-muted-foreground/70 size-4 shrink-0"
          aria-hidden={true}
        />
        <div className="min-w-0">
          <p className="text-foreground truncate text-sm font-medium">
            {file.name || "Untitled"}
          </p>
          <p className="text-muted-foreground text-xs">
            {file.mime_type} · {formatBytes(file.size)}
          </p>
        </div>
      </div>
      <ExternalLink
        className="text-muted-foreground/60 size-3.5 shrink-0"
        aria-hidden={true}
      />
    </button>
  );
}

function ResourceRow({ resource }: { resource: ResourceSummary }) {
  const Icon =
    resource.type === "video"
      ? Video
      : resource.type === "pdf" || resource.type === "article"
        ? FileText
        : ExternalLink;
  return (
    <a
      href={resource.url}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        "bg-card border-border hover:border-foreground/20 flex items-start gap-3 rounded-[10px] border p-3 transition-all",
        "hover:shadow-sm",
      )}
    >
      <Icon
        className="text-muted-foreground/70 mt-0.5 size-4 shrink-0"
        aria-hidden={true}
      />
      <div className="min-w-0 flex-1">
        <p className="text-foreground line-clamp-2 text-sm font-medium">
          {resource.title}
        </p>
        {resource.description ? (
          <p className="text-muted-foreground mt-0.5 line-clamp-2 text-xs">
            {resource.description}
          </p>
        ) : null}
      </div>
    </a>
  );
}
