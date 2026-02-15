import { Hero } from "@/lib/features/marketing/landing/hero";
import { FeaturesDisplay } from "@/lib/features/marketing/landing/features-display";
import { Mission } from "@/lib/features/marketing/landing/mission";
import { CTA } from "@/lib/features/marketing/landing/cta";

export default function MarketingPage() {
  return (
    <>
      <Hero />
      <FeaturesDisplay />
      <Mission />
      <CTA />
    </>
  );
}
