import Link from "next/link";
import { ChevronUp, Globe2, Lock, NotebookPen } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export interface StudyGuideListItemResponse {
  id: string;
  title: string;
  description?: string | null;
  creator: { display_name: string };
  vote_score: number;
  quiz_count: number;
  is_recommended: boolean;
  tags: string[];
  course_id: string;
  // ASK-212: every list item carries the canonical visibility flag.
  // The list response does NOT include a grant count, so we can only
  // distinguish `private` vs `public` here; the "Shared" indicator
  // (private + >=1 grant) surfaces on the detail page instead.
  visibility: "private" | "public";
  /** ISO 8601 timestamp; rendered as a relative "X days ago" label when present. */
  updated_at?: string;
}

interface StudyGuideCardProps {
  guide: StudyGuideListItemResponse;
  variant: "list" | "compact";
  href?: string;
  courseLabel?: string;
}

export function StudyGuideCard({
  guide,
  variant,
  href,
  courseLabel,
}: StudyGuideCardProps) {
  const destination = href ?? `/study-guides/${guide.id}`;

  if (variant === "compact") {
    return (
      <Link
        href={destination}
        className={cn(
          "bg-card border-border block rounded-xl border p-4 transition-all",
          "hover:border-foreground/20 hover:shadow-md",
        )}
      >
        <CompactVariant guide={guide} />
      </Link>
    );
  }

  return (
    <Link
      href={destination}
      className={cn(
        "bg-card border-border group flex h-full flex-col gap-2.5 rounded-[10px] border px-5 pb-4 pt-[18px] transition-all",
        "hover:border-foreground/20 focus-visible:ring-ring hover:shadow-md focus-visible:outline-hidden focus-visible:ring-2",
      )}
    >
      <ListVariant guide={guide} courseLabel={courseLabel} />
    </Link>
  );
}

function CompactVariant({ guide }: { guide: StudyGuideListItemResponse }) {
  return (
    <div className="space-y-1">
      <div className="flex items-start gap-1.5">
        <p className="line-clamp-2 flex-1 text-sm font-medium leading-snug text-foreground">
          {guide.title}
        </p>
        <VisibilityIndicator visibility={guide.visibility} />
      </div>
      <p className="text-xs text-muted-foreground">
        by {guide.creator.display_name} ·{" "}
        {guide.quiz_count === 1 ? "1 quiz" : `${guide.quiz_count} quizzes`}
      </p>
    </div>
  );
}

function ListVariant({
  guide,
  courseLabel,
}: {
  guide: StudyGuideListItemResponse;
  courseLabel?: string;
}) {
  const initial =
    guide.creator.display_name.trim().charAt(0).toUpperCase() || "?";
  const relative = formatRelativeTime(guide.updated_at);

  return (
    <>
      <div className="flex items-start gap-3.5">
        <div className="flex shrink-0 flex-col items-center gap-0.5 pt-0.5">
          <ChevronUp
            className="text-muted-foreground size-3.5"
            aria-hidden={true}
          />
          <span className="text-foreground font-mono text-[12px] font-semibold">
            {guide.vote_score}
          </span>
        </div>
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <div className="flex items-start gap-1.5">
            <h3 className="text-foreground line-clamp-2 flex-1 text-[15px] font-semibold leading-[1.35] tracking-[-0.2px]">
              {guide.title}
            </h3>
            <VisibilityIndicator visibility={guide.visibility} />
            {guide.is_recommended ? (
              <Badge
                variant="secondary"
                className="bg-primary/10 text-primary shrink-0 text-[11px] font-semibold"
              >
                Recommended
              </Badge>
            ) : null}
          </div>
          {guide.description ? (
            <p className="text-muted-foreground line-clamp-2 text-[13px] leading-[1.5]">
              {guide.description}
            </p>
          ) : null}
          {courseLabel ? (
            <p className="text-muted-foreground text-[12px]">{courseLabel}</p>
          ) : null}
          {guide.tags.length > 0 ? (
            <div className="flex flex-wrap gap-1 pt-0.5">
              {guide.tags.map((tag, index) => (
                <Badge
                  key={`${tag}-${index}`}
                  variant="secondary"
                  className="bg-muted text-muted-foreground text-[11px]"
                >
                  {tag}
                </Badge>
              ))}
            </div>
          ) : null}
        </div>
      </div>

      <div className="flex-1" />

      <div className="bg-border/60 h-px w-full" />

      <div className="flex items-center justify-between pt-1.5 text-[12px]">
        <div className="text-muted-foreground flex min-w-0 items-center gap-2">
          <span
            className="bg-muted text-foreground inline-flex size-[22px] shrink-0 items-center justify-center rounded-full font-semibold"
            aria-hidden={true}
          >
            {initial}
          </span>
          <span className="text-foreground/80 truncate font-medium">
            {guide.creator.display_name}
          </span>
          {relative ? (
            <>
              <span className="text-muted-foreground/50" aria-hidden={true}>
                ·
              </span>
              <span className="truncate">{relative}</span>
            </>
          ) : null}
        </div>
        <span className="text-muted-foreground flex shrink-0 items-center gap-1.5 font-medium">
          <NotebookPen
            className="text-muted-foreground/60 size-3"
            aria-hidden={true}
          />
          {guide.quiz_count}
        </span>
      </div>
    </>
  );
}

function formatRelativeTime(iso?: string): string | null {
  if (!iso) return null;
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return null;
  const diffMs = Date.now() - then;
  if (diffMs < 0) return "just now";
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  if (weeks < 4) return `${weeks}w ago`;
  const months = Math.floor(days / 30);
  if (months < 12) return `${months}mo ago`;
  const years = Math.floor(days / 365);
  return `${years}y ago`;
}

function VisibilityIndicator({
  visibility,
}: {
  visibility: StudyGuideListItemResponse["visibility"];
}) {
  const { Icon, label } =
    visibility === "public"
      ? { Icon: Globe2, label: "Public" }
      : { Icon: Lock, label: "Private" };
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            role="img"
            aria-label={label}
            className="text-muted-foreground mt-0.5 inline-flex shrink-0"
          >
            <Icon className="size-3.5" aria-hidden />
          </span>
        </TooltipTrigger>
        <TooltipContent>{label}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
