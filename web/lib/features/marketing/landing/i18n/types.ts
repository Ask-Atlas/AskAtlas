export type LandingFeatureId =
  | "study-guides"
  | "practice-questions"
  | "resources"
  | "community";

export type LandingPillarId = "accessible" | "empower" | "innovate";

interface FeatureCopy {
  title: string;
  description: string;
  displayDescription: string;
  ctaText: string;
}

interface PillarCopy {
  title: string;
  description: string;
}

export interface LandingDictionary {
  hero: {
    badge: string;
    headlineStart: string;
    headlineHighlight: string;
    body: string;
    primaryCta: string;
    secondaryCta: string;
  };
  socialProof: {
    studentsValue: string;
    studentsLabel: string;
    classesValue: string;
    classesLabel: string;
  };
  marquee: {
    heading: string;
  };
  sections: {
    features: {
      headingStart: string;
      headingHighlight: string;
      body: string;
    };
    mission: {
      eyebrow: string;
      headingStart: string;
      headingHighlight: string;
      body: string;
      metricValue: string;
      metricLabelLine1: string;
      metricLabelLine2: string;
    };
    cta: {
      headingStart: string;
      headingHighlight: string;
      body: string;
      primaryCta: string;
      secondaryCta: string;
    };
  };
  features: Record<LandingFeatureId, FeatureCopy>;
  pillars: Record<LandingPillarId, PillarCopy>;
  carousel: {
    active: string;
    inactive: string;
    fallbackCta: string;
  };
}
