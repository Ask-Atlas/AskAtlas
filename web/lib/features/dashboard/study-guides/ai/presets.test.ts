import { AI_EDIT_PRESETS, findPreset, isPresetEligible } from "./presets";

describe("AI_EDIT_PRESETS", () => {
  it("ships the six chips the ticket calls out", () => {
    const ids = AI_EDIT_PRESETS.map((p) => p.id).sort();
    expect(ids).toEqual(
      [
        "add-example",
        "clearer",
        "exam-audience",
        "reorganize",
        "shorten",
        "tldr",
      ].sort(),
    );
  });

  it("preset ids are unique", () => {
    const ids = AI_EDIT_PRESETS.map((p) => p.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("every preset has a non-empty instruction", () => {
    for (const p of AI_EDIT_PRESETS) {
      expect(p.instruction.trim().length).toBeGreaterThan(0);
    }
  });
});

describe("isPresetEligible", () => {
  it("multi-paragraph presets are disabled on single-block selections", () => {
    const reorganize = findPreset("reorganize")!;
    expect(
      isPresetEligible(reorganize, { selectionSpansMultipleBlocks: false }),
    ).toBe(false);
    expect(
      isPresetEligible(reorganize, { selectionSpansMultipleBlocks: true }),
    ).toBe(true);
  });

  it("always-eligible presets are enabled regardless of selection shape", () => {
    const shorten = findPreset("shorten")!;
    expect(
      isPresetEligible(shorten, { selectionSpansMultipleBlocks: false }),
    ).toBe(true);
    expect(
      isPresetEligible(shorten, { selectionSpansMultipleBlocks: true }),
    ).toBe(true);
  });
});

describe("TL;DR transformReplacement", () => {
  it("prepends a TL;DR blockquote to the original", () => {
    const tldr = findPreset("tldr")!;
    expect(tldr.transformReplacement).toBeDefined();
    const out = tldr.transformReplacement!(
      "BSTs keep keys ordered for fast lookup.",
      "A binary search tree (BST) is an ordered tree...",
    );
    expect(out).toBe(
      "> TL;DR: BSTs keep keys ordered for fast lookup.\n\nA binary search tree (BST) is an ordered tree...",
    );
  });

  it("trims whitespace from the AI response", () => {
    const tldr = findPreset("tldr")!;
    const out = tldr.transformReplacement!(
      "  Sloppy whitespace.  \n",
      "Body text.",
    );
    expect(out.startsWith("> TL;DR: Sloppy whitespace.")).toBe(true);
    expect(out).toContain("\n\nBody text.");
  });
});

describe("findPreset", () => {
  it("returns undefined for unknown ids", () => {
    expect(findPreset("does-not-exist")).toBeUndefined();
  });

  it("returns the matching preset for known ids", () => {
    expect(findPreset("tldr")?.id).toBe("tldr");
  });
});
