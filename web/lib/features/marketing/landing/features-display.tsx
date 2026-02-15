"use client";

import { motion } from "motion/react";
import { Section } from "@/lib/features/marketing/section";
import { FeatureDisplayCard } from "./feature-display-card";
import { getFeatures } from "./content-mappers";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function FeaturesDisplay() {
  const copy = useLandingCopy();
  const features = getFeatures(copy);

  return (
    <Section
      id="features"
      className="bg-background"
      background={
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 h-[600px] w-[600px] rounded-full bg-primary/10 blur-[120px] lg:h-[900px] lg:w-[900px]" />
      }
    >
      {/* Section header */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 0.6 }}
        className="mx-auto mb-16 max-w-2xl text-center"
      >
        <h2 className="mb-4 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
          {copy.sections.features.headingStart}{" "}
          <span className="bg-linear-to-r from-primary to-chart-1 bg-clip-text text-transparent">
            {copy.sections.features.headingHighlight}
          </span>
        </h2>
        <p className="text-lg leading-relaxed text-muted-foreground">
          {copy.sections.features.body}
        </p>
      </motion.div>

      {/* Bento grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {features.map((feature, index) => (
          <FeatureDisplayCard
            key={feature.id}
            feature={feature}
            index={index}
          />
        ))}
      </div>
    </Section>
  );
}
