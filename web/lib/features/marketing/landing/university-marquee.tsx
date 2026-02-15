"use client";

import {
  Marquee,
  MarqueeContent,
  MarqueeFade,
  MarqueeItem,
} from "@/components/ui/marquee";
import { UNIVERSITIES } from "./fixtures";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function UniversityMarquee() {
  const copy = useLandingCopy();

  return (
    <div className="w-full px-4 pt-8 lg:px-8">
      <div className="flex items-center justify-center mb-8">
        <span className="text-xs font-semibold tracking-widest text-muted-foreground/60 uppercase">
          {copy.marquee.heading}
        </span>
      </div>
      <Marquee>
        <MarqueeFade side="left" />
        <MarqueeContent speed={40} pauseOnHover>
          {UNIVERSITIES.map((university) => (
            <MarqueeItem key={university}>
              <div className="flex items-center gap-3">
                <span className="text-lg font-bold text-foreground/50 whitespace-nowrap">
                  {university}
                </span>
                <span className="text-primary/20">|</span>
              </div>
            </MarqueeItem>
          ))}
        </MarqueeContent>
        <MarqueeFade side="right" />
      </Marquee>
    </div>
  );
}
