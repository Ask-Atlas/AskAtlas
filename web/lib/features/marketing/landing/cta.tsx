"use client";

import { ArrowRight } from "lucide-react";
import { motion } from "motion/react";
import { Button } from "@/components/ui/button";
import { Section } from "@/lib/features/marketing/section";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function CTA() {
  const copy = useLandingCopy();

  return (
    <Section
      background={
        <>
          {/* Base gradient */}
          <div className="absolute inset-0 bg-linear-to-br from-primary/10 via-background to-chart-2/10" />

          {/* Blob 1 — drifts left-to-right across the top */}
          <motion.div
            className="absolute top-0 h-[350px] w-[350px] rounded-full bg-primary/20 blur-[120px]"
            animate={{ x: ["-20%", "120%"], y: ["-10%", "30%", "-10%"] }}
            transition={{ duration: 14, repeat: Infinity, ease: "easeInOut" }}
          />

          {/* Blob 2 — drifts right-to-left across the bottom */}
          <motion.div
            className="absolute bottom-0 h-[300px] w-[300px] rounded-full bg-chart-2/20 blur-[120px]"
            animate={{ x: ["120%", "-20%"], y: ["10%", "-25%", "10%"] }}
            transition={{ duration: 16, repeat: Infinity, ease: "easeInOut" }}
          />

          {/* Blob 3 — smaller accent, diagonal path */}
          <motion.div
            className="absolute h-[200px] w-[200px] rounded-full bg-chart-1/15 blur-[100px]"
            animate={{
              x: ["60%", "-10%", "60%"],
              y: ["-20%", "110%", "-20%"],
            }}
            transition={{ duration: 18, repeat: Infinity, ease: "easeInOut" }}
          />
        </>
      }
    >
      <motion.div
        initial={{ opacity: 0, y: 40 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.7 }}
        className="text-center"
      >
        <motion.h2
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.6, delay: 0.1 }}
          className="mx-auto mb-6 max-w-2xl text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
        >
          {copy.sections.cta.headingStart}{" "}
          <span className="bg-linear-to-r from-primary to-chart-1 bg-clip-text text-transparent">
            {copy.sections.cta.headingHighlight}
          </span>
        </motion.h2>
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.6, delay: 0.25 }}
          className="mx-auto mb-10 max-w-xl text-lg leading-relaxed text-muted-foreground"
        >
          {copy.sections.cta.body}
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.6, delay: 0.4 }}
          className="flex flex-col items-center justify-center gap-4 sm:flex-row"
        >
          <Button size="lg" className="group px-8 shadow-lg shadow-primary/25">
            {copy.sections.cta.primaryCta}
            <ArrowRight className="ml-2 h-5 w-5 transition-transform group-hover:translate-x-1" />
          </Button>
          <Button size="lg" variant="outline" className="px-8">
            {copy.sections.cta.secondaryCta}
          </Button>
        </motion.div>
      </motion.div>
    </Section>
  );
}
