"use client";

import { cn } from "@/lib/utils";
import { motion } from "motion/react";
import Image from "next/image";
import type { Feature } from "./types";

interface FeatureDisplayCardProps {
  feature: Feature;
  index: number;
}

export function FeatureDisplayCard({
  feature,
  index,
}: FeatureDisplayCardProps) {
  const Icon = feature.icon;
  const imageBottom = feature.layout === "image-bottom";

  const assetArea = (
    <div
      className={cn(
        "relative overflow-hidden rounded-xl bg-linear-to-br",
        feature.color,
        feature.span === "large" ? "h-48" : "h-36",
      )}
    >
      <Image
        src={feature.image}
        alt={feature.title}
        fill
        className="object-cover opacity-90 transition-transform duration-500 hover:scale-105"
        sizes="(max-width: 768px) 100vw, 50vw"
      />
    </div>
  );

  const textArea = (
    <div>
      {/* Animated icon badge */}
      <motion.div
        className={cn(
          "mb-3 inline-flex h-10 w-10 items-center justify-center rounded-lg bg-linear-to-br",
          feature.color,
        )}
        whileHover={{ scale: 1.2, rotate: -12 }}
        transition={{ type: "spring", stiffness: 400, damping: 12 }}
      >
        <Icon className="h-5 w-5 text-white" />
      </motion.div>

      <h3 className="mb-2 text-lg font-semibold text-card-foreground">
        {feature.title}
      </h3>
      <p className="text-sm leading-relaxed text-muted-foreground">
        {feature.displayDescription ?? feature.description}
      </p>
    </div>
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-60px" }}
      transition={{ duration: 0.5, delay: index * 0.1, ease: "easeOut" }}
      whileHover={{ y: -6 }}
      className={cn(
        "group relative overflow-hidden rounded-2xl border border-border bg-card/60 backdrop-blur-sm p-6 transition-shadow duration-300 hover:shadow-xl hover:shadow-primary/5",
        feature.span === "large" && "md:col-span-2",
      )}
    >
      {/* Subtle gradient glow on hover */}
      <div className="absolute inset-0 opacity-0 transition-opacity duration-500 group-hover:opacity-100 bg-linear-to-br from-primary/5 via-transparent to-chart-2/5" />

      <div
        className={cn(
          "relative z-10 flex h-full flex-col gap-6",
          imageBottom && "flex-col-reverse",
        )}
      >
        {assetArea}
        {textArea}
      </div>
    </motion.div>
  );
}
