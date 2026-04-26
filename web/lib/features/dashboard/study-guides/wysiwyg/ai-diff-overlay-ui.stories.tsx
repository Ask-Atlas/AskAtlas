import type { Editor } from "@tiptap/core";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { useEffect, useState } from "react";
import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { AiDiffMark } from "./ai-diff-mark";
import { AiDiffOverlay, getAiDiffState } from "./ai-diff-overlay";
import { AiDiffOverlayUi } from "./ai-diff-overlay-ui";

const SAMPLE_ORIGINAL =
  "The quick brown fox jumps over the lazy dog. It is a classic pangram used to test typewriters.";
const SAMPLE_REPLACEMENT =
  "A fast brown fox leaps over the sleepy dog. It's a classic pangram everyone uses to test fonts.";

interface PlaygroundProps {
  original?: string;
  replacement?: string;
  editId?: string;
}

function findRangeForText(editor: Editor, needle: string) {
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

function Playground({
  original = SAMPLE_ORIGINAL,
  replacement = SAMPLE_REPLACEMENT,
  editId = "11111111-2222-3333-4444-555555555555",
}: PlaygroundProps) {
  const [resolution, setResolution] = useState<null | {
    accepted: boolean;
    finalText: string;
  }>(null);
  const [seeded, setSeeded] = useState(false);

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [StarterKit, AiDiffMark, AiDiffOverlay],
    content: `<p>${original}</p>`,
    editorProps: {
      attributes: {
        class:
          "prose prose-neutral dark:prose-invert max-w-none rounded-md border bg-background p-3 focus:outline-none",
      },
    },
  });

  /* eslint-disable react-hooks/set-state-in-effect --
   * Storybook playground reseeds the editor when args change; the
   * setState calls are the response to that external transition.
   */
  useEffect(() => {
    if (!editor) return;
    setSeeded(false);
    setResolution(null);
    editor.commands.setContent(`<p>${original}</p>`);
    const range = findRangeForText(editor, original);
    if (!range) return;
    const ok = editor.commands.seedAiDiff({
      editId,
      originalText: original,
      replacement,
      from: range.from,
      to: range.to,
    });
    if (ok) setSeeded(true);
  }, [editor, original, replacement, editId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  if (!editor) return <div>Loading editor…</div>;

  const handleResolved = (accepted: boolean) => {
    setResolution({ accepted, finalText: editor.state.doc.textContent });
  };

  const pluginState = seeded ? getAiDiffState(editor.state) : null;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6">
      <section>
        <h3 className="text-muted-foreground mb-2 text-xs font-medium uppercase tracking-wide">
          Editor (with diff marks)
        </h3>
        <EditorContent editor={editor} />
      </section>

      <section>
        <h3 className="text-muted-foreground mb-2 text-xs font-medium uppercase tracking-wide">
          Diff overlay popover body
        </h3>
        <div className="bg-popover w-fit rounded-md border p-3 shadow-md">
          {pluginState ? (
            <AiDiffOverlayUi editor={editor} onResolved={handleResolved} />
          ) : (
            <p className="text-muted-foreground text-xs">
              No diff seeded (replacement equals original).
            </p>
          )}
        </div>
      </section>

      {resolution ? (
        <section className="border-border bg-muted/30 rounded-md border p-3 text-xs">
          <p className="font-medium">
            Resolved · accepted = {String(resolution.accepted)}
          </p>
          <pre className="mt-1 whitespace-pre-wrap font-mono">
            {resolution.finalText}
          </pre>
        </section>
      ) : null}
    </div>
  );
}

const meta: Meta<typeof Playground> = {
  title: "Dashboard/StudyGuides/AiDiffOverlay (ASK-217)",
  component: Playground,
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "Per-hunk accept/reject UI for AI edits. Boots a real TipTap editor + the AiDiffOverlay extension, seeds a diff between two strings, and renders both the editor (with ins/del marks inline) and the popover body. Click ✓/✗ on a hunk row, or Accept all / Reject all -- the editor updates in real time. The resolution panel below echoes the final outcome that ASK-217 PATCHes back to /api/study-guides/{id}/ai/edits/{edit_id}.",
      },
    },
  },
};

export default meta;

type Story = StoryObj<typeof Playground>;

export const ManyHunks: Story = {
  name: "Many hunks (mixed insert/delete/replace)",
};

export const SingleReplace: Story = {
  name: "Single replace",
  args: {
    original: "the quick fox",
    replacement: "the slow fox",
  },
};

export const PureInsert: Story = {
  name: "Pure insertion",
  args: {
    original: "Hello",
    replacement: "Hello, world!",
  },
};

export const PureDelete: Story = {
  name: "Pure deletion",
  args: {
    original: "Hello, world!",
    replacement: "Hello",
  },
};

export const NoChanges: Story = {
  name: "No changes (original equals replacement)",
  args: {
    original: "Nothing to review here.",
    replacement: "Nothing to review here.",
  },
  parameters: {
    docs: {
      description: {
        story:
          "When the AI returns identical text to the original, `seedAiDiff` returns false and the popover body renders nothing -- the bubble menu integration treats this as a no-op resolution.",
      },
    },
  },
};
