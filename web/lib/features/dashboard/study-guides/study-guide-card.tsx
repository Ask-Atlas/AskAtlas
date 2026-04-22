import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { ThumbsUp } from "lucide-react";
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
  const title =
    guide.title.length > 100 ? guide.title.slice(0, 100) + "…" : guide.title;

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
        <CompactVariant title={title} guide={guide} />
      ) : (
        <ListVariant title={title} guide={guide} courseLabel={courseLabel} />
      )}
    </Link>
  );
}

function CompactVariant({
  title,
  guide,
}: {
  title: string;
  guide: StudyGuideListItemResponse;
}) {
  return (
    <div className="space-y-1">
      <p className="text-sm font-medium leading-snug text-neutral-800 dark:text-white">
        {title}
      </p>
      <p className="text-xs text-neutral-500 dark:text-neutral-400">
        by {guide.creator.display_name} · {guide.quiz_count} quizzes
      </p>
    </div>
  );
}

function ListVariant({
  title,
  guide,
  courseLabel,
}: {
  title: string;
  guide: StudyGuideListItemResponse;
  courseLabel?: string;
}) {
  return (
    <div className="space-y-3">
      <div className="flex items-start justify-between gap-2">
        <p className="text-sm font-bold leading-snug text-neutral-800 dark:text-white">
          {title}
        </p>
        {guide.is_recommended && (
          <Badge
            variant="secondary"
            className="shrink-0 text-xs bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400"
          >
            Recommended
          </Badge>
        )}
      </div>

      <p className="text-xs text-neutral-500 dark:text-neutral-400">
        by {guide.creator.display_name}
        {courseLabel ? ` · ${courseLabel}` : ""}
      </p>

      <div className="flex items-center gap-2">
        <Badge
          variant="outline"
          className="flex items-center gap-1 text-xs border-black/10 dark:border-white/20 text-neutral-600 dark:text-neutral-300"
        >
          <ThumbsUp className="h-3 w-3" />
          {guide.vote_score}
        </Badge>
        <Badge
          variant="outline"
          className="text-xs border-black/10 dark:border-white/20 text-neutral-600 dark:text-neutral-300"
        >
          {guide.quiz_count} quizzes
        </Badge>
      </div>

      {guide.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {guide.tags.map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="text-xs bg-black/5 text-neutral-600 dark:bg-white/10 dark:text-neutral-300"
            >
              {tag}
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
