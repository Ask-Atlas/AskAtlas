/**
 * @jest-environment jsdom
 */
import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";

import { AiDiffMark } from "./ai-diff-mark";
import {
  AiDiffOverlay,
  aiDiffPluginKey,
  getAiDiffState,
} from "./ai-diff-overlay";

interface BootArgs {
  initialHtml?: string;
}

function boot({ initialHtml = "<p></p>" }: BootArgs = {}) {
  const dom = document.createElement("div");
  document.body.appendChild(dom);
  const editor = new Editor({
    element: dom,
    extensions: [StarterKit, AiDiffMark, AiDiffOverlay],
    content: initialHtml,
  });
  return { editor, dom };
}

function htmlOf(editor: Editor): string {
  return editor.getHTML();
}

function plainText(editor: Editor): string {
  return editor.state.doc.textContent;
}

function getRangeForText(
  editor: Editor,
  needle: string,
): { from: number; to: number } | null {
  let foundFrom = -1;
  editor.state.doc.descendants((node, pos) => {
    if (foundFrom !== -1) return false;
    if (!node.isText) return;
    const nodeText = node.text ?? "";
    const i = nodeText.indexOf(needle);
    if (i === -1) return;
    foundFrom = pos + i;
  });
  if (foundFrom === -1) return null;
  return { from: foundFrom, to: foundFrom + needle.length };
}

describe("AiDiffOverlay extension", () => {
  it("seedAiDiff replaces the range and seeds plugin state", () => {
    const { editor } = boot({ initialHtml: "<p>the quick fox</p>" });
    const range = getRangeForText(editor, "the quick fox")!;

    const ok = editor.commands.seedAiDiff({
      editId: "edit-1",
      originalText: "the quick fox",
      replacement: "the slow fox",
      from: range.from,
      to: range.to,
    });
    expect(ok).toBe(true);

    const state = getAiDiffState(editor.state);
    expect(state).not.toBeNull();
    expect(state!.editId).toBe("edit-1");
    expect(state!.hunks.length).toBeGreaterThanOrEqual(1);

    const html = htmlOf(editor);
    expect(html).toContain('data-ai-diff="ins"');
    expect(html).toContain('data-ai-diff="del"');
    expect(plainText(editor)).toContain("quick");
    expect(plainText(editor)).toContain("slow");

    editor.destroy();
  });

  it("seedAiDiff returns false when there's nothing to review", () => {
    const { editor } = boot({ initialHtml: "<p>hello</p>" });
    const range = getRangeForText(editor, "hello")!;

    const ok = editor.commands.seedAiDiff({
      editId: "edit-noop",
      originalText: "hello",
      replacement: "hello",
      from: range.from,
      to: range.to,
    });
    expect(ok).toBe(false);
    expect(aiDiffPluginKey.getState(editor.state)).toBeNull();

    editor.destroy();
  });

  it("acceptAiHunk drops del text and strips ins mark for that hunk", () => {
    const { editor } = boot({ initialHtml: "<p>the quick fox</p>" });
    const range = getRangeForText(editor, "the quick fox")!;
    editor.commands.seedAiDiff({
      editId: "edit-2",
      originalText: "the quick fox",
      replacement: "the slow fox",
      from: range.from,
      to: range.to,
    });
    const before = getAiDiffState(editor.state)!;
    const hunkId = before.hunks[0].id;

    const ok = editor.commands.acceptAiHunk(hunkId);
    expect(ok).toBe(true);

    const after = getAiDiffState(editor.state)!;
    expect(after.hunks.find((h) => h.id === hunkId)).toBeUndefined();

    expect(plainText(editor)).toBe("the slow fox");
    const html = htmlOf(editor);
    expect(html).not.toContain(`data-hunk-id="${hunkId}"`);

    editor.destroy();
  });

  it("rejectAiHunk drops ins text and strips del mark for that hunk", () => {
    const { editor } = boot({ initialHtml: "<p>the quick fox</p>" });
    const range = getRangeForText(editor, "the quick fox")!;
    editor.commands.seedAiDiff({
      editId: "edit-3",
      originalText: "the quick fox",
      replacement: "the slow fox",
      from: range.from,
      to: range.to,
    });
    const before = getAiDiffState(editor.state)!;
    const hunkId = before.hunks[0].id;

    const ok = editor.commands.rejectAiHunk(hunkId);
    expect(ok).toBe(true);

    const after = getAiDiffState(editor.state)!;
    expect(after.hunks.find((h) => h.id === hunkId)).toBeUndefined();
    expect(plainText(editor)).toBe("the quick fox");

    editor.destroy();
  });

  it("acceptAllAiHunks resolves to the full replacement", () => {
    const { editor } = boot({
      initialHtml: "<p>the quick brown fox</p>",
    });
    const range = getRangeForText(editor, "the quick brown fox")!;
    editor.commands.seedAiDiff({
      editId: "edit-4",
      originalText: "the quick brown fox",
      replacement: "a fast brown cat",
      from: range.from,
      to: range.to,
    });
    expect(getAiDiffState(editor.state)!.hunks.length).toBeGreaterThan(1);

    editor.commands.acceptAllAiHunks();

    expect(getAiDiffState(editor.state)!.hunks).toHaveLength(0);
    expect(plainText(editor)).toBe("a fast brown cat");

    editor.destroy();
  });

  it("rejectAllAiHunks resolves back to the original", () => {
    const { editor } = boot({
      initialHtml: "<p>the quick brown fox</p>",
    });
    const range = getRangeForText(editor, "the quick brown fox")!;
    editor.commands.seedAiDiff({
      editId: "edit-5",
      originalText: "the quick brown fox",
      replacement: "a fast brown cat",
      from: range.from,
      to: range.to,
    });

    editor.commands.rejectAllAiHunks();

    expect(getAiDiffState(editor.state)!.hunks).toHaveLength(0);
    expect(plainText(editor)).toBe("the quick brown fox");

    editor.destroy();
  });

  it("clearAiDiff drops del-marked text and clears plugin state", () => {
    const { editor } = boot({ initialHtml: "<p>the quick fox</p>" });
    const range = getRangeForText(editor, "the quick fox")!;
    editor.commands.seedAiDiff({
      editId: "edit-6",
      originalText: "the quick fox",
      replacement: "the slow fox",
      from: range.from,
      to: range.to,
    });

    editor.commands.clearAiDiff();
    expect(aiDiffPluginKey.getState(editor.state)).toBeNull();

    const html = htmlOf(editor);
    expect(html).not.toContain('data-ai-diff="del"');

    editor.destroy();
  });

  it("hunk positions stay correct after concurrent typing elsewhere", () => {
    const { editor } = boot({
      initialHtml: "<p>head text</p><p>the quick fox</p>",
    });
    const range = getRangeForText(editor, "the quick fox")!;
    editor.commands.seedAiDiff({
      editId: "edit-7",
      originalText: "the quick fox",
      replacement: "the slow fox",
      from: range.from,
      to: range.to,
    });
    const hunkId = getAiDiffState(editor.state)!.hunks[0].id;

    const headRange = getRangeForText(editor, "head text")!;
    editor.chain().focus().insertContentAt(headRange.from, "PREFIX ").run();

    editor.commands.acceptAiHunk(hunkId);
    expect(plainText(editor)).toContain("the slow fox");
    expect(plainText(editor)).toContain("PREFIX ");

    editor.destroy();
  });
});
