"use client";

import { Sparkles } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function HeroBadge() {
  const copy = useLandingCopy();

  return (
    <div className="relative inline-flex">
      <div className="absolute -inset-0.5 bg-linear-to-r from-primary via-purple-500 to-primary rounded-full opacity-75 blur animate-pulse" />
      <Badge
        variant="secondary"
        className="relative inline-flex items-center gap-2 px-4 py-2 bg-background border-primary/50 text-primary"
      >
        <Sparkles className="h-4 w-4" />
        <span>{copy.hero.badge}</span>
      </Badge>
    </div>
  );
}
