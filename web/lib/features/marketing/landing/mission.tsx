"use client";

import { motion } from "motion/react";
import { Section } from "@/lib/features/marketing/section";
import { MissionPillar } from "./mission-pillar";
import { getPillars } from "./content-mappers";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function Mission() {
  const copy = useLandingCopy();
  const pillars = getPillars(copy);

  return (
    <Section
      className="bg-background"
      background={
        <div className="absolute inset-0 bg-[radial-gradient(circle,var(--color-primary)_1px,transparent_1px)] bg-size-[20px_20px] opacity-[0.04]" />
      }
    >
      <div className="grid items-center gap-16 lg:grid-cols-2">
        {/* Left  */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.6 }}
          className="space-y-6"
        >
          <p className="text-sm font-semibold uppercase tracking-widest text-primary">
            {copy.sections.mission.eyebrow}
          </p>
          <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl lg:leading-[1.15]">
            {copy.sections.mission.headingStart}{" "}
            <span className="bg-linear-to-r from-chart-2 to-primary bg-clip-text text-transparent">
              {copy.sections.mission.headingHighlight}
            </span>
          </h2>
          <p className="max-w-lg text-lg leading-relaxed text-muted-foreground">
            {copy.sections.mission.body}
          </p>

          {/* Highlight metric card */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="mt-8 inline-flex items-center gap-4 rounded-2xl border border-primary/20 bg-primary/5 px-6 py-4 backdrop-blur-sm"
          >
            <span className="text-4xl font-bold text-primary">
              {copy.sections.mission.metricValue}
            </span>
            <span className="text-sm leading-snug text-muted-foreground">
              {copy.sections.mission.metricLabelLine1}
              <br />
              {copy.sections.mission.metricLabelLine2}
            </span>
          </motion.div>
        </motion.div>

        {/* Right — pillar cards */}
        <div className="space-y-5">
          {pillars.map((pillar, index) => (
            <MissionPillar key={pillar.id} pillar={pillar} index={index} />
          ))}
        </div>
      </div>
    </Section>
  );
}
