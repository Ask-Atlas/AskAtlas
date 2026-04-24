import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

interface SkeletonListProps {
  count?: number;
  className?: string;
}

export function SkeletonList({ count = 3, className }: SkeletonListProps) {
  return (
    <ul
      aria-hidden
      data-testid="skeleton-list"
      className={cn("flex flex-col", className)}
    >
      {Array.from({ length: count }).map((_, i) => (
        <li key={i} className="flex items-center gap-3 p-3">
          <Skeleton className="size-10 rounded-md" />
          <div className="flex flex-col gap-2">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-3 w-32" />
          </div>
        </li>
      ))}
    </ul>
  );
}
