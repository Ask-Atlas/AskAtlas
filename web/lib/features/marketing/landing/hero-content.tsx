import { HeroBadge } from "./hero-badge";
import { HeroHeadline } from "./hero-headline";
import { HeroActions } from "./hero-actions";
import { HeroSocialProof } from "./hero-social-proof";

export function HeroContent() {
  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-1000">
      <HeroBadge />
      <HeroHeadline />
      <div className="space-y-8">
        <HeroActions />
        <HeroSocialProof />
      </div>
    </div>
  );
}
