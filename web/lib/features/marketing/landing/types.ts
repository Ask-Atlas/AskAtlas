import type { LucideIcon } from "lucide-react";

export interface Feature {
  id: string;
  title: string;
  description: string;
  icon: LucideIcon;
  image: string;
  color: string;
  ctaText?: string;
  ctaLink?: string;
  /** Extended description for the feature display section. */
  displayDescription?: string;
  span?: "large" | "default";
  layout?: "image-top" | "image-bottom";
}

export interface Pillar {
  id: string;
  title: string;
  description: string;
  icon: LucideIcon;
}
