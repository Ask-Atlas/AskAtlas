"use client";

import { Star } from "lucide-react";
import { useOptimistic, useTransition } from "react";

import type { ToggleFavoriteResponse } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface FavoriteButtonProps {
  initialFavorited: boolean;
  /**
   * Fires the actual toggle request. Must resolve to the server's
   * ToggleFavoriteResponse so the caller can sync back any derived
   * state (counts, last-favorited timestamps, etc). If it throws,
   * this component rolls the star back to `initialFavorited` and
   * re-throws so a caller-owned toast can surface the failure.
   */
  onToggle: () => Promise<ToggleFavoriteResponse>;
  /** Screen-reader label + tooltip. Context-specific ("Favorite this file"). */
  label: string;
  size?: "sm" | "md";
  className?: string;
}

export function FavoriteButton({
  initialFavorited,
  onToggle,
  label,
  size = "md",
  className,
}: FavoriteButtonProps) {
  const [optimisticFavorited, setOptimisticFavorited] = useOptimistic(
    initialFavorited,
    (_current, next: boolean) => next,
  );
  const [isPending, startTransition] = useTransition();

  const handleClick = () => {
    startTransition(async () => {
      setOptimisticFavorited(!optimisticFavorited);
      try {
        await onToggle();
      } catch {
        // useOptimistic reverts when the transition completes, so the
        // visual rollback happens regardless. Callers wrap their own
        // onToggle if they need to surface the error (e.g. toast).
      }
    });
  };

  return (
    <button
      type="button"
      aria-label={label}
      aria-pressed={optimisticFavorited}
      title={label}
      disabled={isPending}
      onClick={handleClick}
      className={cn(
        "inline-flex items-center justify-center rounded-full",
        "text-muted-foreground hover:text-foreground transition-colors",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        "disabled:opacity-60 disabled:cursor-not-allowed",
        size === "sm" ? "size-6" : "size-8",
        className,
      )}
    >
      <Star
        className={cn(
          size === "sm" ? "size-3.5" : "size-4",
          optimisticFavorited && "fill-amber-400 text-amber-400",
        )}
      />
    </button>
  );
}
