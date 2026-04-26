/**
 * Preset edit actions surfaced as chips in the Ask AI popover
 * (ASK-218). Each preset is a canned `instruction` string the
 * existing AI edit endpoint already understands -- no backend work.
 *
 * Tuning lives here: change a label or a prompt and the chip row
 * picks it up on next mount. Keep instructions short + imperative;
 * the edit endpoint's system prompt already establishes the
 * "rewrite only the selection, preserve markdown" contract.
 */

export type PresetEligibility = "always" | "multi-paragraph";

export interface AiEditPreset {
  /** Stable identifier; shipped with future analytics events. */
  id: string;
  /** Short button label. */
  label: string;
  /** Sent to the AI edit endpoint as `instruction`. */
  instruction: string;
  /**
   * When the chip is enabled. `multi-paragraph` chips disable
   * themselves on single-block selections per the ticket spec
   * ("Reorganize disabled on single-paragraph selections").
   */
  eligibility: PresetEligibility;
  /**
   * Optional client-side post-processing applied to the AI's
   * replacement BEFORE seeding the diff overlay. TL;DR uses this
   * to prepend the summary to the original selection rather than
   * replacing it outright (per ticket spec).
   */
  transformReplacement?: (aiResponse: string, originalText: string) => string;
}

export const AI_EDIT_PRESETS: ReadonlyArray<AiEditPreset> = [
  {
    id: "tldr",
    label: "TL;DR",
    instruction:
      "Write a concise TL;DR summary of the selected text. Output ONLY the summary -- no preface, no markdown headers. 1-3 sentences max.",
    eligibility: "always",
    transformReplacement: (aiResponse, original) =>
      `> TL;DR: ${aiResponse.trim()}\n\n${original}`,
  },
  {
    id: "clearer",
    label: "Make clearer",
    instruction:
      "Rewrite the selected text to be clearer and easier to follow. Preserve meaning and markdown formatting.",
    eligibility: "always",
  },
  {
    id: "shorten",
    label: "Shorten",
    instruction:
      "Shorten the selected text by trimming redundant phrasing. Preserve meaning and markdown formatting.",
    eligibility: "always",
  },
  {
    id: "add-example",
    label: "Add an example",
    instruction:
      "Add a single concrete example that illustrates the selected text. Keep the original text intact and append the example after it.",
    eligibility: "always",
  },
  {
    id: "reorganize",
    label: "Reorganize",
    instruction:
      "Reorganize the selected text into a more logical order. Preserve all content and markdown structure.",
    eligibility: "multi-paragraph",
  },
  {
    id: "exam-audience",
    label: "For exam",
    instruction:
      "Rewrite the selected text in a study-friendly tone for an exam audience. Use active voice, define jargon inline, and keep the markdown structure.",
    eligibility: "always",
  },
];

/**
 * Returns whether a preset's chip should be enabled given the
 * current selection. Stays a pure function so tests don't need a
 * real ProseMirror editor.
 */
export function isPresetEligible(
  preset: AiEditPreset,
  ctx: { selectionSpansMultipleBlocks: boolean },
): boolean {
  if (preset.eligibility === "multi-paragraph") {
    return ctx.selectionSpansMultipleBlocks;
  }
  return true;
}

/**
 * Look up a preset by id. Returns undefined for typos / removed ids.
 */
export function findPreset(id: string): AiEditPreset | undefined {
  return AI_EDIT_PRESETS.find((p) => p.id === id);
}
