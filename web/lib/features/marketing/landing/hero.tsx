import { HeroContent } from "./hero-content";
import { FeatureCarousel } from "./feature-carousel";
import { UniversityMarquee } from "./university-marquee";
import { Section } from "@/lib/features/marketing/section";

export function Hero() {
  return (
    <>
      <Section
        className="bg-background"
        background={
          <>
            <div className="absolute -top-24 -right-24 h-[500px] w-[500px] rounded-full bg-primary/20 blur-3xl lg:h-[800px] lg:w-[800px] lg:opacity-30" />
            <div className="absolute -bottom-24 -left-24 h-[400px] w-[400px] rounded-full bg-accent/20 blur-3xl lg:h-[600px] lg:w-[600px] lg:opacity-30" />
            <div className="absolute inset-0 bg-[linear-gradient(to_right,var(--color-primary)_1px,transparent_1px),linear-gradient(to_bottom,var(--color-primary)_1px,transparent_1px)] bg-size-[24px_34px] mask-[radial-gradient(ellipse_80%_50%_at_50%_0%,#000_70%,transparent_100%)] opacity-10" />
          </>
        }
      >
        <div className="grid gap-12 lg:grid-cols-2 lg:gap-16 items-center">
          <HeroContent />
          <FeatureCarousel />
        </div>
      </Section>
      <UniversityMarquee />
    </>
  );
}
