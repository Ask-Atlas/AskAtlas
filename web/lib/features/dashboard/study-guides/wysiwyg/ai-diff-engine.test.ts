import { applyAcceptedHunks, computeHunks } from "./ai-diff-engine";

describe("computeHunks", () => {
  it("returns no hunks for identical strings", () => {
    expect(computeHunks("hello world", "hello world")).toEqual([]);
  });

  it("returns no hunks for two empty strings", () => {
    expect(computeHunks("", "")).toEqual([]);
  });

  it("emits a single replace hunk for a one-word swap", () => {
    const hunks = computeHunks("the quick fox", "the slow fox");
    expect(hunks).toHaveLength(1);
    expect(hunks[0].op).toBe("replace");
    expect(hunks[0].delText).toBe("quick");
    expect(hunks[0].insText).toBe("slow");
    expect(hunks[0].originalStart).toBe(4);
    expect(hunks[0].replacementStart).toBe(4);
  });

  it("emits a pure insert when only chars are added", () => {
    const hunks = computeHunks("hello", "hello world");
    expect(hunks).toHaveLength(1);
    expect(hunks[0]).toMatchObject({
      op: "insert",
      delText: "",
      insText: " world",
      originalStart: 5,
      replacementStart: 5,
    });
  });

  it("emits a pure delete when only chars are removed", () => {
    const hunks = computeHunks("hello world", "world");
    expect(hunks).toHaveLength(1);
    expect(hunks[0].op).toBe("delete");
    expect(hunks[0].delText.trim()).toBe("hello");
  });

  it("groups multiple changes into separate hunks", () => {
    const hunks = computeHunks(
      "the quick brown fox jumps over the lazy dog",
      "a fast brown fox leaps over the sleepy dog",
    );
    expect(hunks.length).toBeGreaterThan(1);
    for (const h of hunks) {
      expect(h.id).toMatch(/^h\d+$/);
    }
    const ids = new Set(hunks.map((h) => h.id));
    expect(ids.size).toBe(hunks.length);
  });

  it("offsets are monotonically increasing in the original string", () => {
    const hunks = computeHunks(
      "alpha bravo charlie delta",
      "ALPHA bravo CHARLIE delta",
    );
    let prev = -1;
    for (const h of hunks) {
      expect(h.originalStart).toBeGreaterThan(prev);
      prev = h.originalStart;
    }
  });
});

describe("applyAcceptedHunks", () => {
  it("returns the original when no hunks are accepted", () => {
    const original = "the quick fox";
    const hunks = computeHunks(original, "the slow fox");
    expect(applyAcceptedHunks(original, hunks, new Set())).toBe(original);
  });

  it("returns the full replacement when all hunks are accepted", () => {
    const original = "the quick fox";
    const replacement = "the slow fox";
    const hunks = computeHunks(original, replacement);
    const allIds = new Set(hunks.map((h) => h.id));
    expect(applyAcceptedHunks(original, hunks, allIds)).toBe(replacement);
  });

  it("applies a mix of accepted + rejected hunks", () => {
    const original = "the quick brown fox jumps over the lazy dog";
    const replacement = "a fast brown fox leaps over the sleepy dog";
    const hunks = computeHunks(original, replacement);
    expect(hunks.length).toBeGreaterThan(1);

    const onlyFirst = new Set([hunks[0].id]);
    const result = applyAcceptedHunks(original, hunks, onlyFirst);

    expect(result).toContain(hunks[0].insText);
    for (const h of hunks.slice(1)) {
      if (h.delText) {
        expect(result).toContain(h.delText);
      }
    }
  });

  it("is a no-op when no hunks exist", () => {
    expect(applyAcceptedHunks("same", [], new Set())).toBe("same");
  });
});
