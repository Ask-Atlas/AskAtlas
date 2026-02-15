"use client";

import { useLandingCopy } from "./i18n/landing-copy-provider";

export function HeroHeadline() {
  const copy = useLandingCopy();

  return (
    <div className="space-y-4 max-w-3xl">
      <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl">
        {copy.hero.headlineStart}{" "}
        <span className="bg-linear-to-r from-primary to-chart-2 bg-clip-text text-transparent">
          {copy.hero.headlineHighlight}
        </span>
      </h1>
      <p className="text-lg text-muted-foreground max-w-xl leading-relaxed">
        {copy.hero.body}
      </p>
    </div>
  );
}
