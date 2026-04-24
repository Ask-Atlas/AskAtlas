"use client";

import Link from "next/link";
import Markdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkGfm from "remark-gfm";

import { cn } from "@/lib/utils";

import { ArticleImage, isInternalHref } from "./article-internals";

interface ArticleRendererProps {
  content: string;
  className?: string;
}

// Extends the default hast-util-sanitize schema to permit the author
// affordances study guides rely on (inline callouts, figure captions).
const sanitizeSchema = {
  ...defaultSchema,
  tagNames: [
    ...(defaultSchema.tagNames ?? []),
    "figure",
    "figcaption",
    "aside",
  ],
  attributes: {
    ...defaultSchema.attributes,
    "*": [...(defaultSchema.attributes?.["*"] ?? []), "className"],
  },
};

/**
 * Renders study-guide markdown as styled HTML. GFM tables + task lists,
 * inline HTML via rehype-raw behind rehype-sanitize's XSS-safe schema,
 * Next.js Link for internal hrefs, external hrefs open in a new tab.
 *
 * Whitespace-only content renders nothing (no stray wrapper).
 */
export function ArticleRenderer({ content, className }: ArticleRendererProps) {
  if (content.trim() === "") return null;

  return (
    <div
      className={cn(
        "prose prose-neutral dark:prose-invert max-w-none",
        className,
      )}
    >
      <Markdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw, [rehypeSanitize, sanitizeSchema]]}
        components={{
          a({ href, children, ...props }) {
            if (typeof href === "string" && isInternalHref(href)) {
              return <Link href={href}>{children}</Link>;
            }
            return (
              <a
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                {...props}
              >
                {children}
              </a>
            );
          },
          img: ({ src, alt }) => (
            <ArticleImage
              src={typeof src === "string" ? src : undefined}
              alt={typeof alt === "string" ? alt : undefined}
            />
          ),
        }}
      >
        {content}
      </Markdown>
    </div>
  );
}
