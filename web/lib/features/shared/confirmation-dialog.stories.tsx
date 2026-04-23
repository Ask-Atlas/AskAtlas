import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { fn } from "storybook/test";

import { Button } from "@/components/ui/button";

import { ConfirmationDialog } from "./confirmation-dialog";

const meta: Meta<typeof ConfirmationDialog> = {
  title: "Shared/ConfirmationDialog",
  component: ConfirmationDialog,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Controlled wrapper around shadcn AlertDialog for destructive-action confirmations. The caller owns open state and close semantics so async `onConfirm` flows can keep the dialog open while awaiting the promise.",
      },
    },
  },
  args: {
    title: "Delete file?",
    description: "This can't be undone. The file will be permanently removed.",
    confirmLabel: "Delete",
    cancelLabel: "Cancel",
    destructive: true,
    disabled: false,
    onConfirm: fn(),
    onOpenChange: fn(),
  },
  argTypes: {
    destructive: { control: "boolean" },
    disabled: { control: "boolean" },
  },
};

export default meta;
type Story = StoryObj<typeof ConfirmationDialog>;

export const Destructive: Story = {
  render: (args) => {
    const [open, setOpen] = useState(false);
    return (
      <div className="flex flex-col items-center gap-4">
        <Button variant="destructive" onClick={() => setOpen(true)}>
          Delete file
        </Button>
        <ConfirmationDialog
          {...args}
          open={open}
          onOpenChange={(next) => {
            setOpen(next);
            args.onOpenChange?.(next);
          }}
          onConfirm={() => {
            args.onConfirm();
            setOpen(false);
          }}
        />
      </div>
    );
  },
};

export const Default: Story = {
  args: {
    destructive: false,
    title: "Save changes?",
    description: "Your changes will be applied immediately.",
    confirmLabel: "Save",
  },
  render: (args) => {
    const [open, setOpen] = useState(false);
    return (
      <div className="flex flex-col items-center gap-4">
        <Button onClick={() => setOpen(true)}>Save changes</Button>
        <ConfirmationDialog
          {...args}
          open={open}
          onOpenChange={(next) => {
            setOpen(next);
            args.onOpenChange?.(next);
          }}
          onConfirm={() => {
            args.onConfirm();
            setOpen(false);
          }}
        />
      </div>
    );
  },
};

export const AsyncConfirm: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "The dialog stays open while the async `onConfirm` runs, so callers can show a loading state (via `disabled`) without the primitive unmounting mid-request.",
      },
    },
  },
  render: (args) => {
    const [open, setOpen] = useState(false);
    const [pending, setPending] = useState(false);
    return (
      <div className="flex flex-col items-center gap-4">
        <Button variant="destructive" onClick={() => setOpen(true)}>
          Delete file (async)
        </Button>
        <ConfirmationDialog
          {...args}
          open={open}
          disabled={pending}
          onOpenChange={setOpen}
          onConfirm={async () => {
            setPending(true);
            await new Promise((r) => setTimeout(r, 1500));
            setPending(false);
            setOpen(false);
          }}
          confirmLabel={pending ? "Deleting..." : "Delete"}
        />
      </div>
    );
  },
};

export const AlwaysOpen: Story = {
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        story: "Forced open for visual review.",
      },
    },
  },
  args: { open: true },
};
