import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { SectionMembershipButton } from "./section-membership-button";

function resolveAfter(ms: number): () => Promise<void> {
  return () => new Promise((resolve) => setTimeout(resolve, ms));
}

function rejectAfter(ms: number): () => Promise<void> {
  return () =>
    new Promise((_resolve, reject) =>
      setTimeout(() => reject(new Error("network")), ms),
    );
}

const meta: Meta<typeof SectionMembershipButton> = {
  title: "Dashboard/SectionMembershipButton",
  component: SectionMembershipButton,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Three-state inline enrollment control. Join is a one-tap optimistic action; Leave goes through the shared ConfirmationDialog so users can't drop a section by accident. The caller drives the `membership` prop after a successful request; failures revert automatically via `useOptimistic` and the caller surfaces a toast.",
      },
    },
  },
  argTypes: {
    membership: {
      control: { type: "radio" },
      options: ["member", "not-member", "unknown"],
    },
  },
};

export default meta;
type Story = StoryObj<typeof SectionMembershipButton>;

export const NotMember: Story = {
  args: {
    membership: "not-member",
    onJoin: resolveAfter(500),
    onLeave: resolveAfter(500),
  },
};

export const Member: Story = {
  args: {
    membership: "member",
    onJoin: resolveAfter(500),
    onLeave: resolveAfter(500),
  },
};

export const Unknown: Story = {
  args: {
    membership: "unknown",
    onJoin: resolveAfter(500),
    onLeave: resolveAfter(500),
  },
};

export const SlowJoin: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Simulates a 2s round-trip so you can see the pending/disabled state after the optimistic swap.",
      },
    },
  },
  args: {
    membership: "not-member",
    onJoin: resolveAfter(2000),
    onLeave: resolveAfter(500),
  },
};

export const JoinFails: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "`onJoin` rejects after ~800ms. The button optimistically shows Enrolled, then reverts to Join when the transition settles. Consumers wrap their own onJoin with a try/catch + toast.",
      },
    },
  },
  args: {
    membership: "not-member",
    onJoin: rejectAfter(800),
    onLeave: resolveAfter(500),
  },
};

export const LeaveFlow: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Start at `member` and click Enrolled to open the leave confirmation. The 1.2s `onLeave` keeps the dialog open + disabled while resolving so the confirm button can show 'Leaving…'.",
      },
    },
  },
  args: {
    membership: "member",
    onJoin: resolveAfter(500),
    onLeave: resolveAfter(1200),
  },
};
