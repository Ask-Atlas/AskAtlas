"use client";

import { CardBody, CardContainer, CardItem } from "@/components/ui/3d-card";
import { Button } from "@/components/ui/button";
import Image from "next/image";
import type { Feature } from "./types";
import { cn } from "@/lib/utils";
import { useLandingCopy } from "./i18n/landing-copy-provider";

interface FeatureCardProps {
  feature: Feature;
  isActive: boolean;
}

export function FeatureCard({ feature, isActive }: FeatureCardProps) {
  const Icon = feature.icon;
  const copy = useLandingCopy();

  return (
    <CardContainer className="inter-var">
      <CardBody
        className={cn(
          "bg-gray-50 relative group/card dark:hover:shadow-2xl dark:hover:shadow-emerald-500/10 dark:bg-black dark:border-white/20 border-black/10 w-auto sm:w-[30rem] h-auto rounded-xl p-6 border transition-all duration-500",
          !isActive && "opacity-50",
        )}
      >
        {/* Icon badge */}
        <CardItem translateZ="50" className="mb-4">
          <div
            className={cn(
              "inline-flex items-center justify-center w-12 h-12 rounded-lg bg-linear-to-r",
              feature.color,
            )}
          >
            <Icon className="h-6 w-6 text-white" />
          </div>
        </CardItem>

        {/* Title */}
        <CardItem
          translateZ="50"
          className="text-xl font-bold text-neutral-600 dark:text-white"
        >
          {feature.title}
        </CardItem>

        {/* Description */}
        <CardItem
          as="p"
          translateZ="60"
          className="text-neutral-500 text-sm max-w-sm mt-2 dark:text-neutral-300"
        >
          {feature.description}
        </CardItem>

        <CardItem translateZ="100" className="w-full mt-4">
          <div
            className={cn(
              "relative w-full aspect-video rounded-xl overflow-hidden",
              "group-hover/card:shadow-xl transition-shadow",
            )}
          >
            <Image
              src={feature.image}
              alt={feature.title}
              fill
              className="object-cover"
              sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
            />
          </div>
        </CardItem>

        {/* CTA */}
        <div className="flex justify-between items-center mt-8">
          <CardItem
            translateZ={20}
            as="p"
            className="text-xs text-neutral-500 dark:text-neutral-400"
          >
            {isActive ? copy.carousel.active : copy.carousel.inactive}
          </CardItem>
          <CardItem translateZ={20}>
            <Button
              size="sm"
              className={cn(
                "bg-linear-to-r text-white font-bold",
                feature.color,
              )}
            >
              {feature.ctaText || copy.carousel.fallbackCta} →
            </Button>
          </CardItem>
        </div>
      </CardBody>
    </CardContainer>
  );
}
