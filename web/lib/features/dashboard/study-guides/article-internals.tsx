"use client";

import { useState } from "react";

import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

/**
 * Internal URLs go through Next.js Link; protocol-relative URLs (//)
 * are intentionally treated as external so a pasted `//evil.example`
 * can't masquerade as same-origin.
 */
export function isInternalHref(href: string): boolean {
  return href.startsWith("/") && !href.startsWith("//");
}

interface ArticleImageProps {
  src?: string;
  alt?: string;
}

/**
 * Image with a skeleton placeholder while loading and a graceful
 * fallback on error. Alt text doubles as a figcaption when present.
 */
export function ArticleImage({ src, alt }: ArticleImageProps) {
  const [state, setState] = useState<"loading" | "loaded" | "error">("loading");
  if (!src) return null;

  return (
    <figure className="my-4">
      {state === "loading" ? (
        <Skeleton
          className="aspect-video w-full"
          data-testid="article-image-skeleton"
        />
      ) : null}
      {state === "error" ? (
        <div
          role="alert"
          className="text-muted-foreground rounded-md border border-dashed p-8 text-center text-sm"
        >
          Image failed to load
        </div>
      ) : null}
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={src}
        alt={alt ?? ""}
        className={cn(
          "w-full rounded-md",
          state === "loaded" ? "block" : "hidden",
        )}
        onLoad={() => setState("loaded")}
        onError={() => setState("error")}
      />
      {alt ? (
        <figcaption className="text-muted-foreground mt-2 text-center text-sm">
          {alt}
        </figcaption>
      ) : null}
    </figure>
  );
}
