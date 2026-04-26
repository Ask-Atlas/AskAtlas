import { Extension } from "@tiptap/core";
import type { Mark } from "@tiptap/pm/model";
import { Fragment } from "@tiptap/pm/model";
import type { EditorState, Transaction } from "@tiptap/pm/state";
import { Plugin, PluginKey } from "@tiptap/pm/state";

import { computeHunks, type Hunk, type HunkOp } from "./ai-diff-engine";
import { AI_DIFF_MARK_NAME } from "./ai-diff-mark";

/**
 * AI diff overlay (ASK-217). Owns the lifecycle of a per-edit review
 * session inside the editor:
 *
 *   1. seedAiDiff replaces the selection range with a merged fragment
 *      that contains BOTH the original text (marked `del`) and the
 *      AI's replacement (marked `ins`). The user sees both in flow.
 *   2. acceptAiHunk / rejectAiHunk / *All commands resolve hunks by
 *      removing one side and stripping the mark from the other.
 *   3. clearAiDiff drops every aiDiff mark in the doc (used on
 *      dismissal so a half-resolved diff doesn't pollute the
 *      next save).
 *
 * The hunk list in plugin state is purely for the React UI to render
 * a checklist; the authoritative positions live on the marks
 * themselves and ride through `tr.mapping` automatically (which is
 * what makes "concurrent typing elsewhere" not break the overlay).
 */

export interface AiDiffPluginHunk {
  id: string;
  op: HunkOp;
  /**
   * Original text removed by this hunk (empty for pure inserts).
   * Surfaced so the popover can show "<old> -> <new>" instead of
   * an opaque hunk id.
   */
  delText: string;
  /**
   * Replacement text inserted by this hunk (empty for pure deletes).
   */
  insText: string;
}

export interface AiDiffPluginState {
  editId: string | null;
  hunks: AiDiffPluginHunk[];
}

export const aiDiffPluginKey = new PluginKey<AiDiffPluginState | null>(
  "aiDiffOverlay",
);

interface SeedArgs {
  editId: string | null;
  originalText: string;
  replacement: string;
  from: number;
  to: number;
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    aiDiffOverlay: {
      seedAiDiff: (args: SeedArgs) => ReturnType;
      acceptAiHunk: (hunkId: string) => ReturnType;
      rejectAiHunk: (hunkId: string) => ReturnType;
      acceptAllAiHunks: () => ReturnType;
      rejectAllAiHunks: () => ReturnType;
      clearAiDiff: () => ReturnType;
    };
  }
}

export const AiDiffOverlay = Extension.create({
  name: "aiDiffOverlay",

  addProseMirrorPlugins() {
    return [
      new Plugin<AiDiffPluginState | null>({
        key: aiDiffPluginKey,
        state: {
          init: () => null,
          apply(tr, value) {
            const meta = tr.getMeta(aiDiffPluginKey);
            if (meta !== undefined) {
              return meta as AiDiffPluginState | null;
            }
            return value;
          },
        },
      }),
    ];
  },

  addCommands() {
    return {
      seedAiDiff:
        (args: SeedArgs) =>
        ({ state, dispatch, tr }) => {
          const hunks = computeHunks(args.originalText, args.replacement);
          if (hunks.length === 0) return false;
          const fragment = buildMergedFragment(state, args.originalText, hunks);
          if (!fragment) return false;
          if (dispatch) {
            tr.replaceWith(args.from, args.to, fragment);
            tr.setMeta(aiDiffPluginKey, {
              editId: args.editId,
              hunks: hunks.map((h) => ({
                id: h.id,
                op: h.op,
                delText: h.delText,
                insText: h.insText,
              })),
            } satisfies AiDiffPluginState);
            dispatch(tr);
          }
          return true;
        },

      acceptAiHunk:
        (hunkId: string) =>
        ({ state, dispatch, tr }) => {
          const before = aiDiffPluginKey.getState(state);
          if (!before) return false;
          if (!resolveHunkInTr(state, tr, hunkId, "accept")) return false;
          if (dispatch) {
            tr.setMeta(aiDiffPluginKey, {
              ...before,
              hunks: before.hunks.filter((h) => h.id !== hunkId),
            } satisfies AiDiffPluginState);
            dispatch(tr);
          }
          return true;
        },

      rejectAiHunk:
        (hunkId: string) =>
        ({ state, dispatch, tr }) => {
          const before = aiDiffPluginKey.getState(state);
          if (!before) return false;
          if (!resolveHunkInTr(state, tr, hunkId, "reject")) return false;
          if (dispatch) {
            tr.setMeta(aiDiffPluginKey, {
              ...before,
              hunks: before.hunks.filter((h) => h.id !== hunkId),
            } satisfies AiDiffPluginState);
            dispatch(tr);
          }
          return true;
        },

      acceptAllAiHunks:
        () =>
        ({ state, dispatch, tr }) => {
          const before = aiDiffPluginKey.getState(state);
          if (!before || before.hunks.length === 0) return false;
          for (const h of before.hunks) {
            resolveHunkInTr(state, tr, h.id, "accept");
          }
          if (dispatch) {
            tr.setMeta(aiDiffPluginKey, {
              ...before,
              hunks: [],
            } satisfies AiDiffPluginState);
            dispatch(tr);
          }
          return true;
        },

      rejectAllAiHunks:
        () =>
        ({ state, dispatch, tr }) => {
          const before = aiDiffPluginKey.getState(state);
          if (!before || before.hunks.length === 0) return false;
          for (const h of before.hunks) {
            resolveHunkInTr(state, tr, h.id, "reject");
          }
          if (dispatch) {
            tr.setMeta(aiDiffPluginKey, {
              ...before,
              hunks: [],
            } satisfies AiDiffPluginState);
            dispatch(tr);
          }
          return true;
        },

      clearAiDiff:
        () =>
        ({ state, dispatch, tr }) => {
          const markType = state.schema.marks[AI_DIFF_MARK_NAME];
          if (!markType) return false;
          // Drop any del-marked text the user never accepted, then
          // strip every aiDiff mark off whatever's left. Used on
          // dismiss-without-resolution.
          const delRanges = collectMarkRanges(state, null, "del");
          for (let i = delRanges.length - 1; i >= 0; i--) {
            const r = delRanges[i];
            tr.delete(tr.mapping.map(r.from), tr.mapping.map(r.to));
          }
          tr.removeMark(0, tr.doc.content.size, markType);
          tr.setMeta(aiDiffPluginKey, null);
          if (dispatch) dispatch(tr);
          return true;
        },
    };
  },
});

// --- helpers --------------------------------------------------------

interface MarkRange {
  from: number;
  to: number;
  op: "ins" | "del";
  hunkId: string;
}

function collectMarkRanges(
  state: EditorState,
  hunkId: string | null,
  op: "ins" | "del" | null,
): MarkRange[] {
  const out: MarkRange[] = [];
  state.doc.descendants((node, pos) => {
    if (!node.isText) return;
    for (const mark of node.marks) {
      if (mark.type.name !== AI_DIFF_MARK_NAME) continue;
      const markOp = mark.attrs.op as "ins" | "del";
      const markHunk = mark.attrs.hunkId as string;
      if (hunkId !== null && markHunk !== hunkId) continue;
      if (op !== null && markOp !== op) continue;
      out.push({
        from: pos,
        to: pos + node.nodeSize,
        op: markOp,
        hunkId: markHunk,
      });
    }
  });
  out.sort((a, b) => a.from - b.from);
  return out;
}

function resolveHunkInTr(
  state: EditorState,
  tr: Transaction,
  hunkId: string,
  decision: "accept" | "reject",
): boolean {
  const markType = state.schema.marks[AI_DIFF_MARK_NAME];
  if (!markType) return false;
  const ranges = collectMarkRanges(state, hunkId, null);
  if (ranges.length === 0) return false;

  const dropOp = decision === "accept" ? "del" : "ins";
  const keepOp = decision === "accept" ? "ins" : "del";

  // Walk in reverse so earlier positions don't get shifted by later
  // deletes.
  for (let i = ranges.length - 1; i >= 0; i--) {
    const r = ranges[i];
    if (r.op !== dropOp) continue;
    tr.delete(tr.mapping.map(r.from), tr.mapping.map(r.to));
  }
  // Strip the mark off the surviving side so it renders as plain text.
  for (const r of ranges) {
    if (r.op !== keepOp) continue;
    tr.removeMark(tr.mapping.map(r.from), tr.mapping.map(r.to), markType);
  }
  return true;
}

function buildMergedFragment(
  state: EditorState,
  originalText: string,
  hunks: Hunk[],
): Fragment | null {
  const markType = state.schema.marks[AI_DIFF_MARK_NAME];
  const hardBreak = state.schema.nodes.hardBreak;
  if (!markType) return null;

  // Build (text, mark|null) segments covering the whole transition.
  // Equal text appears once (no mark); ins and del appear with their
  // respective marks + hunkId so per-hunk commands can find them.
  const segments: Array<{ text: string; mark: Mark | null }> = [];
  let originalCursor = 0;

  function pushEqualUntil(originalEnd: number) {
    if (originalCursor < originalEnd) {
      segments.push({
        text: originalText.slice(originalCursor, originalEnd),
        mark: null,
      });
      originalCursor = originalEnd;
    }
  }

  for (const hunk of hunks) {
    pushEqualUntil(hunk.originalStart);
    if (hunk.delText) {
      segments.push({
        text: hunk.delText,
        mark: markType.create({ op: "del", hunkId: hunk.id }),
      });
      originalCursor += hunk.delText.length;
    }
    if (hunk.insText) {
      segments.push({
        text: hunk.insText,
        mark: markType.create({ op: "ins", hunkId: hunk.id }),
      });
    }
  }
  if (originalCursor < originalText.length) {
    segments.push({
      text: originalText.slice(originalCursor),
      mark: null,
    });
  }

  // Convert to inline nodes, splitting on \n into hardBreak nodes so
  // multi-line replacements stay inside the original block.
  const nodes = [];
  for (const seg of segments) {
    const parts = seg.text.split("\n");
    for (let i = 0; i < parts.length; i++) {
      if (i > 0 && hardBreak) {
        nodes.push(hardBreak.create());
      }
      if (parts[i]) {
        nodes.push(state.schema.text(parts[i], seg.mark ? [seg.mark] : []));
      }
    }
  }
  return Fragment.fromArray(nodes);
}

/**
 * Convenience for callers (the React overlay) to read the active
 * diff state. Returns `null` when no review is in progress.
 */
export function getAiDiffState(state: EditorState): AiDiffPluginState | null {
  return aiDiffPluginKey.getState(state) ?? null;
}
