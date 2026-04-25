/**
 * @jest-environment jsdom
 */
import {
  addRecentPrompt,
  getRecentPrompts,
  RECENT_PROMPTS_LIMIT,
  RECENT_PROMPTS_STORAGE_KEY,
} from "./recent-prompts";

describe("recent-prompts", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("returns empty list when storage is empty", () => {
    expect(getRecentPrompts()).toEqual([]);
  });

  it("inserts a new prompt at the head", () => {
    addRecentPrompt("make it shorter");
    expect(getRecentPrompts()).toEqual(["make it shorter"]);
  });

  it("dedupes case-insensitively, moving the duplicate to the head", () => {
    addRecentPrompt("a");
    addRecentPrompt("b");
    addRecentPrompt("A"); // duplicate of "a"
    expect(getRecentPrompts()).toEqual(["A", "b"]);
  });

  it(`caps history at ${RECENT_PROMPTS_LIMIT} entries`, () => {
    for (let i = 0; i < RECENT_PROMPTS_LIMIT + 3; i += 1) {
      addRecentPrompt(`prompt ${i}`);
    }
    const stored = getRecentPrompts();
    expect(stored).toHaveLength(RECENT_PROMPTS_LIMIT);
    // Most-recent-first: last insert is at index 0.
    expect(stored[0]).toBe(`prompt ${RECENT_PROMPTS_LIMIT + 2}`);
  });

  it("ignores empty / whitespace-only prompts", () => {
    addRecentPrompt("real");
    addRecentPrompt("   ");
    addRecentPrompt("");
    expect(getRecentPrompts()).toEqual(["real"]);
  });

  it("trims whitespace before storing", () => {
    addRecentPrompt("  shorten this  ");
    expect(getRecentPrompts()).toEqual(["shorten this"]);
  });

  it("survives a corrupt storage value", () => {
    window.localStorage.setItem(RECENT_PROMPTS_STORAGE_KEY, "{not json");
    expect(getRecentPrompts()).toEqual([]);
    addRecentPrompt("recover");
    expect(getRecentPrompts()).toEqual(["recover"]);
  });

  it("ignores non-array stored values", () => {
    window.localStorage.setItem(
      RECENT_PROMPTS_STORAGE_KEY,
      JSON.stringify({ foo: "bar" }),
    );
    expect(getRecentPrompts()).toEqual([]);
  });

  it("filters out non-string entries from corrupt arrays", () => {
    window.localStorage.setItem(
      RECENT_PROMPTS_STORAGE_KEY,
      JSON.stringify(["valid", 42, null, "also valid"]),
    );
    expect(getRecentPrompts()).toEqual(["valid", "also valid"]);
  });
});
