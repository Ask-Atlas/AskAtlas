import { Mark, mergeAttributes } from "@tiptap/core";

/**
 * Inline mark used to paint AI-edit ins/del runs (ASK-217). Carries
 * a `hunkId` attr so per-hunk accept/reject commands can target the
 * right marks even when several hunks share the same line.
 *
 * Two reasons for a mark instead of a Decoration:
 *   1. Marks ride through ProseMirror's `tr.mapping` automatically,
 *      so concurrent typing elsewhere in the doc doesn't unmoor the
 *      diff annotations.
 *   2. Multi-paragraph selections turn into multiple text nodes;
 *      marks can be re-applied per-paragraph and still share a
 *      hunkId, while a single Decoration would have to be split.
 *
 * `inclusive: false` keeps fresh typing at the boundary of an
 * ins/del run from inheriting the diff styling.
 */

export interface AiDiffMarkAttrs {
  op: "ins" | "del";
  hunkId: string;
}

export const AiDiffMark = Mark.create({
  name: "aiDiff",

  inclusive: false,
  // Excludes itself so a range can't be marked twice. Two distinct
  // hunks on the same character would be a bug in the seeder.
  excludes: "_",

  addAttributes() {
    return {
      op: {
        default: "ins",
        parseHTML: (el) => el.getAttribute("data-ai-diff") ?? "ins",
        renderHTML: (attrs) => ({ "data-ai-diff": attrs.op as string }),
      },
      hunkId: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-hunk-id") ?? "",
        renderHTML: (attrs) => ({ "data-hunk-id": attrs.hunkId as string }),
      },
    };
  },

  parseHTML() {
    return [{ tag: "span[data-ai-diff]" }];
  },

  renderHTML({ HTMLAttributes }) {
    const op = HTMLAttributes["data-ai-diff"] as string | undefined;
    const cls = op === "del" ? "ai-diff-del" : "ai-diff-ins";
    return ["span", mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
});

export const AI_DIFF_MARK_NAME = "aiDiff";
