import DiffMatchPatch from "diff-match-patch";

/**
 * Pure diff engine for ASK-217. Given the AI's `original_span` and
 * `replacement` strings, returns a list of hunks the user can
 * accept or reject individually. Stays decoupled from ProseMirror
 * so it can be unit-tested without spinning up an editor.
 *
 * Hunk model:
 *   - `op === "insert"`  : pure insertion; nothing existed at this
 *                          spot in the original.
 *   - `op === "delete"`  : pure removal; the AI dropped some text
 *                          without putting anything in its place.
 *   - `op === "replace"` : a delete + insert collocated at the same
 *                          equality boundary -- the AI rewrote a span.
 *
 * Offsets (`originalStart`, `replacementStart`) are character indices
 * within the source strings, i.e. `original.slice(0, originalStart)`
 * is the equal text leading up to this hunk. The caller maps these
 * to ProseMirror positions by adding the selection's `from`.
 */

export type HunkOp = "insert" | "delete" | "replace";

export interface Hunk {
  id: string;
  op: HunkOp;
  /** Text removed from the original (empty for pure inserts). */
  delText: string;
  /** Text inserted in the replacement (empty for pure deletes). */
  insText: string;
  /** 0-indexed char offset in `original` where this hunk's removed text starts. */
  originalStart: number;
  /** 0-indexed char offset in `replacement` where this hunk's inserted text starts. */
  replacementStart: number;
}

const DIFF_DELETE = -1;
const DIFF_EQUAL = 0;
const DIFF_INSERT = 1;

/**
 * Computes the user-facing hunks. Returns an empty array when the
 * inputs are identical or both empty.
 */
export function computeHunks(original: string, replacement: string): Hunk[] {
  if (original === replacement) return [];
  const dmp = new DiffMatchPatch();
  const raw = dmp.diff_main(original, replacement);
  // Cleanup-semantic merges trivial single-char shifts back into
  // their surrounding equality so the user sees coherent hunks
  // ("the cat" -> "a cat") instead of every changed character.
  dmp.diff_cleanupSemantic(raw);

  const hunks: Hunk[] = [];
  let originalCursor = 0;
  let replacementCursor = 0;
  let nextId = 0;

  for (let i = 0; i < raw.length; i++) {
    const [op, text] = raw[i];
    if (op === DIFF_EQUAL) {
      originalCursor += text.length;
      replacementCursor += text.length;
      continue;
    }
    if (op === DIFF_DELETE) {
      // Peek at the next op: a DELETE immediately followed by an
      // INSERT is a "replace" hunk -- bundle them so the user sees
      // one decision, not two.
      const next = raw[i + 1];
      if (next && next[0] === DIFF_INSERT) {
        hunks.push({
          id: `h${nextId++}`,
          op: "replace",
          delText: text,
          insText: next[1],
          originalStart: originalCursor,
          replacementStart: replacementCursor,
        });
        originalCursor += text.length;
        replacementCursor += next[1].length;
        i += 1; // consume the INSERT we just paired
        continue;
      }
      hunks.push({
        id: `h${nextId++}`,
        op: "delete",
        delText: text,
        insText: "",
        originalStart: originalCursor,
        replacementStart: replacementCursor,
      });
      originalCursor += text.length;
      continue;
    }
    if (op === DIFF_INSERT) {
      hunks.push({
        id: `h${nextId++}`,
        op: "insert",
        delText: "",
        insText: text,
        originalStart: originalCursor,
        replacementStart: replacementCursor,
      });
      replacementCursor += text.length;
      continue;
    }
  }

  return hunks;
}

/**
 * Applies a subset of hunks to `original`, leaving the rest as the
 * original text. Used when the user "Apply"s after picking a mix of
 * accept/reject decisions: accepted hunks contribute their `insText`,
 * rejected hunks fall back to their `delText` (i.e. the original).
 *
 * Accepted ids are checked once; everything else is treated as
 * rejected (including unresolved hunks at "Apply" time).
 */
export function applyAcceptedHunks(
  original: string,
  hunks: readonly Hunk[],
  acceptedIds: ReadonlySet<string>,
): string {
  if (hunks.length === 0) return original;
  let out = "";
  let cursor = 0;
  for (const hunk of hunks) {
    if (cursor < hunk.originalStart) {
      out += original.slice(cursor, hunk.originalStart);
    }
    if (acceptedIds.has(hunk.id)) {
      out += hunk.insText;
    } else {
      out += hunk.delText;
    }
    cursor = hunk.originalStart + hunk.delText.length;
  }
  if (cursor < original.length) {
    out += original.slice(cursor);
  }
  return out;
}
