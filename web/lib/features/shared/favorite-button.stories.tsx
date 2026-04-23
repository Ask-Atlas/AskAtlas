import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { ToggleFavoriteResponse } from "@/lib/api/types";

import { FavoriteButton } from "./favorite-button";

function okAfter(
  ms: number,
  favorited: boolean,
): () => Promise<ToggleFavoriteResponse> {
  return () =>
    new Promise((resolve) =>
      setTimeout(
        () =>
          resolve({
            favorited,
            favorited_at: favorited ? new Date().toISOString() : null,
          }),
        ms,
      ),
    );
}

function rejectAfter(ms: number): () => Promise<ToggleFavoriteResponse> {
  return () =>
    new Promise((_resolve, reject) =>
      setTimeout(() => reject(new Error("network")), ms),
    );
}

const meta: Meta<typeof FavoriteButton> = {
  title: "Shared/FavoriteButton",
  component: FavoriteButton,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "React 19 `useOptimistic` + `useTransition` star toggle. The star fills the moment the user clicks; the click is blocked while `onToggle` is in flight. Once the transition settles the component reverts to `initialFavorited` -- the caller is expected to re-render with the new value derived from the `ToggleFavoriteResponse`.",
      },
    },
  },
  argTypes: {
    size: {
      control: { type: "radio" },
      options: ["sm", "md"],
    },
  },
  args: {
    label: "Favorite this file",
  },
};

export default meta;
type Story = StoryObj<typeof FavoriteButton>;

export const Unfavorited: Story = {
  args: {
    initialFavorited: false,
    onToggle: okAfter(300, true),
  },
};

export const Favorited: Story = {
  args: {
    initialFavorited: true,
    onToggle: okAfter(300, false),
  },
};

export const SmallSize: Story = {
  args: {
    initialFavorited: true,
    size: "sm",
    onToggle: okAfter(300, false),
  },
};

export const SlowNetwork: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Simulates a 2s server round-trip so you can see the optimistic star plus the disabled/pending state.",
      },
    },
  },
  args: {
    initialFavorited: false,
    onToggle: okAfter(2000, true),
  },
};

export const NetworkFailure: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "`onToggle` rejects -- `useOptimistic` reverts on settle (the component swallows internally, callers layer their own toast).",
      },
    },
  },
  args: {
    initialFavorited: false,
    onToggle: rejectAfter(800),
  },
};
