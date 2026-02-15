"use client";

import { motion } from "motion/react";
import type { Pillar } from "./types";

interface MissionPillarProps {
  pillar: Pillar;
  index: number;
}

export function MissionPillar({ pillar, index }: MissionPillarProps) {
  const Icon = pillar.icon;

  return (
    <motion.div
      initial={{ opacity: 0, x: 40 }}
      whileInView={{ opacity: 1, x: 0 }}
      viewport={{ once: true, margin: "-40px" }}
      transition={{ duration: 0.5, delay: 0.15 * index, ease: "easeOut" }}
      whileHover={{ x: 8 }}
      className="group relative flex items-start gap-5 rounded-xl border border-border/50 bg-card/40 p-5 backdrop-blur-sm transition-colors hover:bg-card/70"
    >
      {/* Accent bar */}
      <div className="absolute left-0 top-0 h-full w-1 rounded-l-xl bg-linear-to-b from-primary to-chart-2 opacity-60 transition-opacity group-hover:opacity-100" />

      {/* Animated icon */}
      <motion.div
        className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-primary/10"
        whileHover={{ scale: 1.15, rotate: -8 }}
        transition={{ type: "spring", stiffness: 400, damping: 12 }}
      >
        <Icon className="h-6 w-6 text-primary" />
      </motion.div>

      <div>
        <h3 className="mb-1 text-base font-semibold text-card-foreground">
          {pillar.title}
        </h3>
        <p className="text-sm leading-relaxed text-muted-foreground">
          {pillar.description}
        </p>
      </div>
    </motion.div>
  );
}
