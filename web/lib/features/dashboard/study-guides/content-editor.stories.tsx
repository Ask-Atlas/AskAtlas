import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { ContentEditor } from "./content-editor";

const SG_ID = "11111111-2222-3333-4444-555555555555";

function Playground({
  initial = "",
  hosts = ["askatlas.app", "localhost"],
}: {
  initial?: string;
  hosts?: string[];
}) {
  const [value, setValue] = useState(initial);
  return (
    <div className="mx-auto max-w-2xl">
      <ContentEditor
        value={value}
        onChange={setValue}
        allowedHosts={hosts}
        rows={14}
        placeholder="Write your study guide in markdown… Paste an askatlas.app/study-guides URL to embed a live card."
      />
      <div className="bg-muted/40 mt-4 rounded border p-3 font-mono text-xs">
        <strong>Source:</strong>
        <pre className="mt-1 whitespace-pre-wrap">{value || "(empty)"}</pre>
      </div>
    </div>
  );
}

const meta: Meta<typeof Playground> = {
  title: "Dashboard/ContentEditor",
  component: Playground,
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "Write + Preview tabs around the content textarea. Pasting an app-internal URL (study-guides / quizzes / files / courses) rewrites it to a directive in the source at the caret; ArticleRenderer in the Preview tab then hydrates it as a live card. Pasting an external URL or a multi-URL clipboard falls through to default paste.",
      },
    },
  },
};

export default meta;
type Story = StoryObj<typeof Playground>;

export const Empty: Story = {};

export const WithContent: Story = {
  args: {
    initial: [
      "# BST primer",
      "",
      "A **binary search tree** keeps keys ordered for O(log n) lookups in the balanced case.",
      "",
      "> Traversal is left-root-right for in-order.",
    ].join("\n"),
  },
};

export const PasteHint: Story = {
  parameters: {
    docs: {
      description: {
        story: `Try it: paste \`https://askatlas.app/study-guides/${SG_ID}\` into the textarea. The editor rewrites it to \`::sg{id="…"}\` at the caret. Flip to Preview to see the hydrated card.`,
      },
    },
  },
  args: {
    initial:
      "## Paste an app URL anywhere below this line — it will rewrite to a directive.\n\n",
  },
};
