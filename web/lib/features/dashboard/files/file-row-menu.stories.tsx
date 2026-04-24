import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { FileResponse } from "@/lib/api/types";

import { FileCard } from "./file-card";
import { FileRowMenu } from "./file-row-menu";

function makeFile(overrides: Partial<FileResponse> = {}): FileResponse {
  return {
    id: "f_preview_1",
    name: "Lecture 3 - Linear Algebra Review.pdf",
    size: 1_048_576,
    mime_type: "application/pdf",
    status: "complete",
    created_at: new Date(Date.now() - 2 * 86_400_000).toISOString(),
    updated_at: new Date(Date.now() - 2 * 86_400_000).toISOString(),
    favorited_at: null,
    last_viewed_at: null,
    ...overrides,
  };
}

function resolveAfter(ms: number): () => Promise<void> {
  return () => new Promise((resolve) => setTimeout(resolve, ms));
}

function rejectAfter(ms: number): () => Promise<void> {
  return () =>
    new Promise((_resolve, reject) =>
      setTimeout(() => reject(new Error("network")), ms),
    );
}

const meta: Meta<typeof FileRowMenu> = {
  title: "Dashboard/FileRowMenu",
  component: FileRowMenu,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Inline rename + delete menu designed to live in the FileCard `rowMenu` slot. Rename swaps the trigger for an auto-focused input (filename pre-selected). Delete opens the shared ConfirmationDialog. While either request is in flight the trigger shows a spinner and the menu disables so callers can't double-submit.",
      },
    },
  },
};

export default meta;
type Story = StoryObj<typeof FileRowMenu>;

export const Default: Story = {
  args: {
    file: makeFile(),
    onRename: () => resolveAfter(500)(),
    onDelete: resolveAfter(500),
  },
};

export const SlowRename: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Rename takes 2s so you can see the disabled input + spinner in the trigger after the swap back.",
      },
    },
  },
  args: {
    file: makeFile(),
    onRename: () => resolveAfter(2000)(),
    onDelete: resolveAfter(500),
  },
};

export const RenameFails: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "`onRename` rejects after ~800ms. The menu returns to the trigger (with the original filename left untouched on `file`). Consumers wrap their own onRename with try/catch + toast.",
      },
    },
  },
  args: {
    file: makeFile(),
    onRename: () => rejectAfter(800)(),
    onDelete: resolveAfter(500),
  },
};

export const SlowDelete: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Confirm the delete dialog to see the 1.5s pending state -- confirm label swaps to 'Deleting…' and both buttons disable.",
      },
    },
  },
  args: {
    file: makeFile(),
    onRename: () => resolveAfter(500)(),
    onDelete: resolveAfter(1500),
  },
};

export const InsideFileCard: Story = {
  parameters: {
    layout: "centered",
    docs: {
      description: {
        story:
          "Shows the canonical placement: FileRowMenu mounted into the FileCard `rowMenu` slot on the list variant.",
      },
    },
  },
  render: (args) => (
    <div className="w-[620px]">
      <FileCard
        file={args.file}
        variant="list"
        rowMenu={
          <FileRowMenu
            file={args.file}
            onRename={args.onRename}
            onDelete={args.onDelete}
          />
        }
      />
    </div>
  ),
  args: {
    file: makeFile(),
    onRename: () => resolveAfter(500)(),
    onDelete: resolveAfter(500),
  },
};
