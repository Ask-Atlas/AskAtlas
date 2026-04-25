import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { AskAiPopover } from "./ask-ai-popover";
import { TiptapEditor } from "./editor";

const SG_ID = "11111111-2222-3333-4444-555555555555";

const SAMPLE = [
  "# Binary search trees",
  "",
  "A **binary search tree** keeps keys ordered for O(log n) lookups in the balanced case.",
  "",
  "Select any sentence above to see the bubble menu — it now exposes inline formatting plus an **Ask AI** entry that opens a prompt input anchored to your selection (ASK-216).",
  "",
  "> The diff overlay (ASK-217) will replace the streaming preview in the popover with a per-hunk accept/reject UI.",
].join("\n");

function Playground({ aiEnabled = true }: { aiEnabled?: boolean }) {
  const [value, setValue] = useState(SAMPLE);
  return (
    <div className="mx-auto w-full max-w-2xl">
      <TiptapEditor
        value={value}
        onChange={setValue}
        allowedHosts={["askatlas.app"]}
        aiEdit={
          aiEnabled
            ? { guideId: SG_ID, title: "BST primer (storybook)" }
            : undefined
        }
      />
      <p className="text-muted-foreground mt-3 text-xs">
        Tip: select text in the document to surface the floating toolbar.
      </p>
    </div>
  );
}

const meta: Meta<typeof Playground> = {
  title: "Dashboard/StudyGuides/EditorBubbleMenu (ASK-216)",
  component: Playground,
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "Selection bubble menu for the study-guide editor. Shows on non-empty text selection (TipTap v3 BubbleMenu, Floating UI). Includes inline formatting toggles (bold / italic / strike / inline code) and -- when an `aiEdit` target is provided -- an **Ask AI** entry that anchors a Radix Popover to the selection. The popover hosts the prompt input, recent-prompts dropdown, and live-streamed preview while the AI edit endpoint streams. Esc and outside-click both cancel cleanly.",
      },
    },
  },
};

export default meta;

type Story = StoryObj<typeof Playground>;

export const FormattingOnly: Story = {
  name: "Formatting only (create-mode)",
  args: { aiEnabled: false },
  parameters: {
    docs: {
      description: {
        story:
          "Create-mode editor: no `aiEdit` prop, so the bubble menu only shows formatting buttons. Used by the new-study-guide page where we don't yet have a guide UUID.",
      },
    },
  },
};

export const WithAskAI: Story = {
  name: "With Ask AI (edit-mode)",
  args: { aiEnabled: true },
  parameters: {
    docs: {
      description: {
        story:
          "Edit-mode editor: passes `aiEdit={{ guideId, title }}`. Selecting text reveals the **Ask AI** entry; clicking it opens the prompt popover. Submitting calls the live API and requires Clerk auth -- in storybook the request will 401, but the popover lifecycle, recent-prompts dropdown, and Esc/outside-click cancel are all interactive.",
      },
    },
  },
};

export const PopoverIdle: Story = {
  name: "Popover · idle",
  parameters: {
    docs: {
      description: {
        story:
          "Standalone view of the AskAiPopover body in its initial state, before the user submits a prompt. The recent-prompts dropdown only renders when there's localStorage history.",
      },
    },
  },
  render: () => (
    <div className="bg-popover w-fit rounded-md border p-3 shadow-md">
      <AskAiPopover
        status="idle"
        replacement=""
        error={null}
        onSubmit={() => undefined}
        onCancel={() => undefined}
      />
    </div>
  ),
};

export const PopoverStreaming: Story = {
  name: "Popover · streaming",
  parameters: {
    docs: {
      description: {
        story:
          "Mid-stream: the input is locked, the cancel button reads `Stop`, and the live preview panel renders the partial replacement with a blinking caret.",
      },
    },
  },
  render: () => (
    <div className="bg-popover w-fit rounded-md border p-3 shadow-md">
      <AskAiPopover
        status="streaming"
        replacement="A binary search tree keeps keys"
        error={null}
        onSubmit={() => undefined}
        onCancel={() => undefined}
      />
    </div>
  ),
};

export const PopoverDone: Story = {
  name: "Popover · done",
  parameters: {
    docs: {
      description: {
        story:
          "After the stream's terminal `done` event. ASK-217's diff overlay will take over from here; today the popover just shows the final replacement and waits to be dismissed.",
      },
    },
  },
  render: () => (
    <div className="bg-popover w-fit rounded-md border p-3 shadow-md">
      <AskAiPopover
        status="done"
        replacement="A binary search tree (BST) maintains an ordered key invariant so lookups, insertions, and deletions all run in O(log n) time when the tree stays balanced."
        error={null}
        onSubmit={() => undefined}
        onCancel={() => undefined}
      />
    </div>
  ),
};

export const PopoverError: Story = {
  name: "Popover · error (429)",
  parameters: {
    docs: {
      description: {
        story:
          "Quota-exceeded / rate-limit error path -- the AppError envelope's `message` is rendered inline so the user can retry without reading the network tab.",
      },
    },
  },
  render: () => (
    <div className="bg-popover w-fit rounded-md border p-3 shadow-md">
      <AskAiPopover
        status="error"
        replacement=""
        error={{
          message: "Daily AI edit quota exceeded. Try again at midnight UTC.",
          status: 429,
        }}
        onSubmit={() => undefined}
        onCancel={() => undefined}
      />
    </div>
  ),
};
