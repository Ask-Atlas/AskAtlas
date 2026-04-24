import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";

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
          "Dropdown menu shell for the FileCard `rowMenu` slot. Exposes **Rename** (which fires `onStartRename` so the caller can flip FileCard into rename mode) and **Delete** (which opens the shared ConfirmationDialog). Rename UX itself lives on FileCard via its `rename` prop -- the input renders in-place of the filename.",
      },
    },
  },
};

export default meta;
type Story = StoryObj<typeof FileRowMenu>;

export const Default: Story = {
  args: {
    file: makeFile(),
    onStartRename: () => {
      // no-op for this story; see InsideFileCard for the real flow
    },
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
    onStartRename: () => {},
    onDelete: resolveAfter(1500),
  },
};

export const DeleteFails: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "`onDelete` rejects after ~800ms. The dialog closes and the trigger returns -- consumers wrap their own onDelete with try/catch + toast.",
      },
    },
  },
  args: {
    file: makeFile(),
    onStartRename: () => {},
    onDelete: rejectAfter(800),
  },
};

export const InsideFileCard: Story = {
  parameters: {
    layout: "centered",
    docs: {
      description: {
        story:
          "Canonical placement + rename flow: the caller tracks 'which file is renaming' in local state, passes `rename={{...}}` to FileCard when that file is active, and wires Rename in the menu via `onStartRename`. Click the three dots and pick Rename -- the filename text swaps to an auto-focused input right where the filename was.",
      },
    },
  },
  render: (args) => <InsideFileCardDemo file={args.file} />,
  args: {
    file: makeFile(),
    onStartRename: () => {},
    onDelete: resolveAfter(500),
  },
};

function InsideFileCardDemo({ file }: { file: FileResponse }) {
  // The caller owns rename-mode state. In a real page this would be
  // keyed by file.id across a list of rows; here we're showing a single
  // row so a boolean is enough.
  const [renaming, setRenaming] = useState(false);
  const [currentName, setCurrentName] = useState(file.name);
  const liveFile: FileResponse = { ...file, name: currentName };

  return (
    <div className="w-[620px]">
      <FileCard
        file={liveFile}
        variant="list"
        rename={
          renaming
            ? {
                onCommit: async (newName) => {
                  // Simulate a 500ms server round-trip.
                  await new Promise((r) => setTimeout(r, 500));
                  setCurrentName(newName);
                  setRenaming(false);
                },
                onCancel: () => setRenaming(false),
              }
            : undefined
        }
        rowMenu={
          <FileRowMenu
            file={liveFile}
            onStartRename={() => setRenaming(true)}
            onDelete={resolveAfter(500)}
          />
        }
      />
    </div>
  );
}
