"use client";

import { Globe2, Lock, Users2 } from "lucide-react";
import type { ReactNode } from "react";

import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import type { StudyGuideVisibility } from "@/lib/api/types";
import { cn } from "@/lib/utils";

/**
 * Footer chip that exposes the current visibility state of a study
 * guide. Clicking opens a popover (`children`) where the caller wires
 * up the Private/Public toggle + grants manager.
 *
 * The "Shared" label surfaces whenever `visibility === "private"` AND
 * the caller passes `grantCount >= 1`; otherwise the label tracks
 * `visibility` verbatim. Keeping the display logic here keeps the
 * form module focused on its react-hook-form wiring.
 */

export interface VisibilityChipProps {
  visibility: StudyGuideVisibility;
  /** Popover surface contents. Usually a toggle + (optional) grants manager. */
  children: ReactNode;
  /**
   * Number of non-creator grants on the guide. When `visibility` is
   * `private` AND this is `>= 1`, the chip shows "Shared · N" instead.
   * Omit / pass 0 in create mode where grants cannot exist yet.
   */
  grantCount?: number;
  disabled?: boolean;
  /** Hidden but helpful for downstream assertions + a11y. */
  id?: string;
}

interface ChipDisplay {
  icon: typeof Lock;
  label: string;
  tooltip: string;
}

function resolveDisplay(
  visibility: StudyGuideVisibility,
  grantCount: number,
): ChipDisplay {
  if (visibility === "public") {
    return {
      icon: Globe2,
      label: "Public",
      tooltip: "Anyone with the link can view",
    };
  }
  if (grantCount >= 1) {
    return {
      icon: Users2,
      label: `Shared · ${grantCount}`,
      tooltip: `Private with ${grantCount} share${grantCount === 1 ? "" : "s"}`,
    };
  }
  return {
    icon: Lock,
    label: "Private",
    tooltip: "Only you can view",
  };
}

export function VisibilityChip({
  visibility,
  children,
  grantCount = 0,
  disabled = false,
  id,
}: VisibilityChipProps) {
  const display = resolveDisplay(visibility, grantCount);
  const Icon = display.icon;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          id={id}
          disabled={disabled}
          aria-label={`Visibility: ${display.label}`}
          title={display.tooltip}
          className={cn(
            "text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-ring",
            "inline-flex h-7 items-center gap-1.5 rounded-full border border-dashed px-3 text-xs",
            "transition-colors focus-visible:outline-none focus-visible:ring-2",
            "disabled:cursor-not-allowed disabled:opacity-60",
          )}
        >
          <Icon className="size-3.5 shrink-0" aria-hidden />
          <span>{display.label}</span>
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-80 space-y-3">
        {children}
      </PopoverContent>
    </Popover>
  );
}
