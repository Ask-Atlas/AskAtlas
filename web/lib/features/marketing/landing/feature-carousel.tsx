"use client";
import * as React from "react";
import Autoplay from "embla-carousel-autoplay";
import {
  Carousel,
  type CarouselApi,
  CarouselContent,
  CarouselItem,
} from "@/components/ui/carousel";
import { cn } from "@/lib/utils";
import { FeatureCard } from "./feature-card";
import { getFeatures } from "./content-mappers";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function FeatureCarousel() {
  const copy = useLandingCopy();
  const features = getFeatures(copy);
  const [api, setApi] = React.useState<CarouselApi>();
  const [current, setCurrent] = React.useState(0);

  const autoplayPlugin = React.useRef(
    Autoplay({ delay: 5000, stopOnInteraction: true, stopOnMouseEnter: true }),
  );

  React.useEffect(() => {
    if (!api) return;

    const handleSelect = () => {
      setCurrent(api.selectedScrollSnap());
    };

    handleSelect();
    api.on("select", handleSelect);

    return () => {
      api.off("select", handleSelect);
    };
  }, [api]);

  return (
    <div className="relative">
      <Carousel
        className="w-full"
        opts={{
          loop: true,
          align: "center",
        }}
        plugins={[autoplayPlugin.current]}
        setApi={setApi}
      >
        <CarouselContent className="-ml-4">
          {features.map((feature, index) => (
            <CarouselItem key={feature.id} className="pl-4">
              <FeatureCard feature={feature} isActive={index === current} />
            </CarouselItem>
          ))}
        </CarouselContent>
      </Carousel>

      {/* Slide indicators */}
      <div className="mt-6 flex justify-center gap-2">
        {features.map((feature, index) => (
          <button
            key={feature.id}
            onClick={() => api?.scrollTo(index)}
            className={cn(
              "h-2 rounded-full transition-all",
              index === current
                ? "w-8 bg-primary"
                : "w-2 bg-gray-300 hover:bg-gray-400 dark:bg-gray-700 dark:hover:bg-gray-600",
            )}
            aria-label={`Go to ${feature.title}`}
            aria-current={index === current ? "true" : undefined}
          />
        ))}
      </div>
    </div>
  );
}
