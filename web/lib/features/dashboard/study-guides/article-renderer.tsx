"use client";

import Link from "next/link";
import { useMemo } from "react";
import Markdown, { type Components } from "react-markdown";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkDirective from "remark-directive";
import remarkGfm from "remark-gfm";

import type { RefSummary } from "@/lib/api/types";
import { cn } from "@/lib/utils";

import { ArticleImage, isInternalHref } from "./article-internals";
import { CalloutBlock } from "./refs/callout-block";
import { CourseRefCard } from "./refs/course-ref-card";
import { EntityRefProvider } from "./refs/entity-ref-context";
import { extractRefs } from "./refs/extract-refs";
import { FileRefCard } from "./refs/file-ref-card";
import { QuizRefCard } from "./refs/quiz-ref-card";
import { remarkAskAtlasDirectives } from "./refs/remark-ask-atlas-directives";
import { StudyGuideRefCard } from "./refs/study-guide-ref-card";

interface ArticleRendererProps {
  content: string;
  className?: string;
  initialRefs?: Record<string, RefSummary | null>;
}

const REF_TAG_NAMES = [
  "ask-sg-ref",
  "ask-quiz-ref",
  "ask-file-ref",
  "ask-course-ref",
  "ask-callout",
] as const;

const sanitizeSchema = {
  ...defaultSchema,
  tagNames: [
    ...(defaultSchema.tagNames ?? []),
    "figure",
    "figcaption",
    "aside",
    ...REF_TAG_NAMES,
  ],
  attributes: {
    ...defaultSchema.attributes,
    "*": [...(defaultSchema.attributes?.["*"] ?? []), "className"],
    "ask-sg-ref": ["id", "dataInline"],
    "ask-quiz-ref": ["id", "dataInline"],
    "ask-file-ref": ["id", "dataInline"],
    "ask-course-ref": ["id", "dataInline"],
    "ask-callout": ["dataCalloutType"],
  },
};

/**
 * Renders study-guide markdown as styled HTML. GFM tables + task lists,
 * inline HTML via rehype-raw behind rehype-sanitize's XSS-safe schema,
 * Next.js Link for internal hrefs, external hrefs open in a new tab.
 *
 * Whitespace-only content renders nothing (no stray wrapper).
 */
export function ArticleRenderer({
  content,
  className,
  initialRefs,
}: ArticleRendererProps) {
  const refs = useMemo(() => extractRefs(content), [content]);

  if (content.trim() === "") return null;

  return (
    <EntityRefProvider refs={refs} initial={initialRefs}>
      <div
        className={cn(
          "prose prose-neutral dark:prose-invert max-w-none",
          className,
        )}
      >
        <Markdown
          remarkPlugins={[remarkGfm, remarkDirective, remarkAskAtlasDirectives]}
          rehypePlugins={[rehypeRaw, [rehypeSanitize, sanitizeSchema]]}
          components={markdownComponents}
        >
          {content}
        </Markdown>
      </div>
    </EntityRefProvider>
  );
}

type RefCardComponent = (props: {
  id: string;
  inline?: boolean;
}) => React.JSX.Element;

function refTag(Card: RefCardComponent) {
  const Rendered = (props: Record<string, unknown>) => {
    const id = typeof props.id === "string" ? props.id : undefined;
    const inline = props["data-inline"] === "1";
    if (!id) return null;
    return <Card id={id} inline={inline} />;
  };
  Rendered.displayName = "RefTag";
  return Rendered;
};

const CalloutTag = (props: Record<string, unknown>) => {
  const type =
    typeof props["data-callout-type"] === "string"
      ? (props["data-callout-type"] as string)
      : undefined;
  return <CalloutBlock type={type}>{props.children as React.ReactNode}</CalloutBlock>;
};

const markdownComponents = {
  a({ href, children, ...props }) {
    if (typeof href === "string" && isInternalHref(href)) {
      return <Link href={href}>{children}</Link>;
    }
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
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
  "ask-sg-ref": refTag(StudyGuideRefCard),
  "ask-quiz-ref": refTag(QuizRefCard),
  "ask-file-ref": refTag(FileRefCard),
  "ask-course-ref": refTag(CourseRefCard),
  "ask-callout": CalloutTag,
} satisfies Components & Record<string, unknown>;
