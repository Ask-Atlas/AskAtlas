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
      <p className="line-clamp-2 text-sm font-medium leading-snug text-foreground">
        {guide.title}
      </p>
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
        <p className="line-clamp-2 text-sm font-bold leading-snug text-foreground">
          {guide.title}
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
