"use client";

import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function HeroActions() {
  const copy = useLandingCopy();

  return (
    <div className="flex flex-col sm:flex-row gap-4">
      <Button size="lg" className="group shadow-lg shadow-primary/20">
        {copy.hero.primaryCta}
        <ArrowRight className="ml-2 h-5 w-5 transition-transform group-hover:translate-x-1" />
      </Button>
      <Button size="lg" variant="outline">
        {copy.hero.secondaryCta}
      </Button>
    </div>
  );
}
