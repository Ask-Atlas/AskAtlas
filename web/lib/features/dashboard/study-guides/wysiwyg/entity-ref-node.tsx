"use client";

import { Node, mergeAttributes } from "@tiptap/core";
import {
  NodeViewWrapper,
  ReactNodeViewRenderer,
  type NodeViewProps,
} from "@tiptap/react";

import { CourseRefCard } from "../refs/course-ref-card";
import { FileRefCard } from "../refs/file-ref-card";
import { QuizRefCard } from "../refs/quiz-ref-card";
import { StudyGuideRefCard } from "../refs/study-guide-ref-card";

export const ENTITY_TYPES = ["sg", "quiz", "file", "course"] as const;
export type EntityType = (typeof ENTITY_TYPES)[number];

const CARDS: Record<
  EntityType,
  (props: { id: string; inline?: boolean }) => React.JSX.Element
> = {
  sg: StudyGuideRefCard,
  quiz: QuizRefCard,
  file: FileRefCard,
  course: CourseRefCard,
};

function EntityRefView({ node }: NodeViewProps) {
  const type = node.attrs.type as EntityType;
  const id = node.attrs.id as string;
  const inline = (node.attrs.variant as string) === "inline";
  const Card = CARDS[type];
  if (!Card || !id) return null;
  return (
    <NodeViewWrapper
      as={inline ? "span" : "div"}
      contentEditable={false}
      className={inline ? "inline-block" : "block"}
      data-entity-ref={type}
    >
      <Card id={id} inline={inline} />
    </NodeViewWrapper>
  );
}

export const EntityRefNode = Node.create({
  name: "entityRef",
  group: "inline",
  inline: true,
  atom: true,
  selectable: true,
  draggable: false,

  addAttributes() {
    return {
      type: { default: "sg" },
      id: { default: "" },
      variant: { default: "leaf" },
    };
  },

  parseHTML() {
    const rules: Array<{
      tag: string;
      getAttrs?: (el: HTMLElement) => Record<string, string>;
    }> = [];
    for (const t of ENTITY_TYPES) {
      rules.push({
        tag: `${t}-ref`,
        getAttrs: (el) => ({
          type: t,
          id: el.getAttribute("id") ?? "",
          variant: "leaf",
        }),
      });
      rules.push({
        tag: `${t}-ref-inline`,
        getAttrs: (el) => ({
          type: t,
          id: el.getAttribute("id") ?? "",
          variant: "inline",
        }),
      });
    }
    return rules;
  },

  renderHTML({ node, HTMLAttributes }) {
    const type = node.attrs.type as string;
    const inline = node.attrs.variant === "inline";
    const tag = inline ? `${type}-ref-inline` : `${type}-ref`;
    return [tag, mergeAttributes({ id: node.attrs.id }, HTMLAttributes)];
  },

  addNodeView() {
    return ReactNodeViewRenderer(EntityRefView);
  },

  addStorage() {
    return {
      markdown: {
        serialize(
          this: unknown,
          state: { write: (s: string) => void },
          node: { attrs: { type: string; id: string; variant: string } },
        ) {
          const prefix = node.attrs.variant === "inline" ? ":" : "::";
          state.write(`${prefix}${node.attrs.type}{id="${node.attrs.id}"}`);
        },
      },
    };
  },
});
