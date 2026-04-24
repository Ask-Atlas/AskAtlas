"use client";

import type { ReactNode } from "react";
import { AlertTriangle, Info, Lightbulb } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export const CALLOUT_TYPES = ["note", "warning", "tip"] as const;
export type CalloutType = (typeof CALLOUT_TYPES)[number];

interface Props {
  type?: string;
  children: ReactNode;
}

interface Variant {
  container: string;
  iconWrap: string;
  Icon: LucideIcon;
  label: string;
}

const VARIANTS: Record<CalloutType, Variant> = {
  note: {
    container: "border-sky-400/40 bg-sky-50 dark:bg-sky-950/40",
    iconWrap: "text-sky-600 dark:text-sky-400",
    Icon: Info,
    label: "Note",
  },
  warning: {
    container: "border-amber-400/40 bg-amber-50 dark:bg-amber-950/40",
    iconWrap: "text-amber-600 dark:text-amber-400",
    Icon: AlertTriangle,
    label: "Warning",
  },
  tip: {
    container: "border-emerald-400/40 bg-emerald-50 dark:bg-emerald-950/40",
    iconWrap: "text-emerald-600 dark:text-emerald-400",
    Icon: Lightbulb,
    label: "Tip",
  },
};

function resolveType(raw?: string): CalloutType {
  if (raw && (CALLOUT_TYPES as readonly string[]).includes(raw)) {
    return raw as CalloutType;
  }
  if (raw && process.env.NODE_ENV !== "production") {
    console.warn(
      `[callout] unknown type ${JSON.stringify(raw)}, falling back to "note"`,
    );
  }
  return "note";
}

export function CalloutBlock({ type, children }: Props) {
  const variant = VARIANTS[resolveType(type)];
  const { Icon, label } = variant;

  return (
    <aside
      role="note"
      aria-label={label}
      className={cn(
        "my-4 flex gap-3 rounded-lg border border-l-4 p-4 text-sm",
        variant.container,
      )}
    >
      <Icon
        className={cn("mt-0.5 size-5 shrink-0", variant.iconWrap)}
        aria-hidden
      />
      <div className="flex-1 [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">
        {children}
      </div>
    </aside>
  );
}
