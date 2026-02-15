import {
  BookOpen,
  BrainCircuit,
  Heart,
  GraduationCap,
  Lightbulb,
  Sparkles,
  Users,
} from "lucide-react";
import type { Feature, Pillar } from "./types";
import type { LandingDictionary } from "./i18n/types";

export function getFeatures(copy: LandingDictionary): Feature[] {
  return [
    {
      id: "study-guides",
      title: copy.features["study-guides"].title,
      description: copy.features["study-guides"].description,
      displayDescription: copy.features["study-guides"].displayDescription,
      icon: BookOpen,
      image: "/features/study-guide.png",
      color: "from-blue-500 to-cyan-500",
      ctaText: copy.features["study-guides"].ctaText,
      span: "large",
      layout: "image-top",
    },
    {
      id: "practice-questions",
      title: copy.features["practice-questions"].title,
      description: copy.features["practice-questions"].description,
      displayDescription:
        copy.features["practice-questions"].displayDescription,
      icon: BrainCircuit,
      image: "/features/practice-questions.png",
      color: "from-purple-500 to-pink-500",
      ctaText: copy.features["practice-questions"].ctaText,
      layout: "image-bottom",
    },
    {
      id: "resources",
      title: copy.features.resources.title,
      description: copy.features.resources.description,
      displayDescription: copy.features.resources.displayDescription,
      icon: Sparkles,
      image: "/features/resources.png",
      color: "from-orange-500 to-red-500",
      ctaText: copy.features.resources.ctaText,
      layout: "image-top",
    },
    {
      id: "community",
      title: copy.features.community.title,
      description: copy.features.community.description,
      displayDescription: copy.features.community.displayDescription,
      icon: Users,
      image: "/features/community.png",
      color: "from-green-500 to-emerald-500",
      ctaText: copy.features.community.ctaText,
      span: "large",
      layout: "image-bottom",
    },
  ];
}

export function getPillars(copy: LandingDictionary): Pillar[] {
  return [
    {
      id: "accessible",
      title: copy.pillars.accessible.title,
      description: copy.pillars.accessible.description,
      icon: Heart,
    },
    {
      id: "empower",
      title: copy.pillars.empower.title,
      description: copy.pillars.empower.description,
      icon: GraduationCap,
    },
    {
      id: "innovate",
      title: copy.pillars.innovate.title,
      description: copy.pillars.innovate.description,
      icon: Lightbulb,
    },
  ];
}
