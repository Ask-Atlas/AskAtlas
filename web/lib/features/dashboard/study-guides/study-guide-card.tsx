import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Globe2, Lock, ThumbsUp } from "lucide-react";
import { cn } from "@/lib/utils";

export interface StudyGuideListItemResponse {
  id: string;
  title: string;
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

  return (
    <Link
      href={destination}
      className={cn(
        "block rounded-xl border p-4 transition-all duration-200",
        "bg-gray-50 dark:bg-black",
        "border-black/10 dark:border-white/20",
        "hover:shadow-md dark:hover:shadow-2xl dark:hover:shadow-emerald-500/10",
      )}
    >
      {variant === "compact" ? (
        <CompactVariant guide={guide} />
      ) : (
        <ListVariant guide={guide} courseLabel={courseLabel} />
      )}
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
  return (
    <div className="space-y-3">
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 flex-1 items-start gap-1.5">
          <p className="line-clamp-2 flex-1 text-sm font-bold leading-snug text-foreground">
            {guide.title}
          </p>
          <VisibilityIndicator visibility={guide.visibility} />
        </div>
        {guide.is_recommended && (
          <Badge
            variant="secondary"
            className="shrink-0 text-xs bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400"
          >
            Recommended
          </Badge>
        )}
      </div>

      <p className="text-xs text-muted-foreground">
        by {guide.creator.display_name}
        {courseLabel ? ` · ${courseLabel}` : ""}
      </p>

      <div className="flex items-center gap-2">
        <Badge
          variant="outline"
          className="flex items-center gap-1 text-xs border-border text-muted-foreground"
        >
          <ThumbsUp className="h-3 w-3" />
          {guide.vote_score}
        </Badge>
        <Badge
          variant="outline"
          className="text-xs border-border text-muted-foreground"
        >
          {guide.quiz_count === 1 ? "1 quiz" : `${guide.quiz_count} quizzes`}
        </Badge>
      </div>

      {guide.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {guide.tags.map((tag, index) => (
            <Badge
              key={`${tag}-${index}`}
              variant="secondary"
              className="text-xs bg-muted text-muted-foreground"
            >
              {tag}
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
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
