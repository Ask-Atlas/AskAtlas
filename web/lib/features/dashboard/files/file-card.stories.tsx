import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { MoreVertical } from "lucide-react";
import { fn } from "storybook/test";

import { Button } from "@/components/ui/button";
import type { FileResponse, ToggleFavoriteResponse } from "@/lib/api/types";
import { FavoriteButton } from "@/lib/features/shared/favorite-button";

import { FileCard } from "./file-card";

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

const mockFavoriteToggle: () => Promise<ToggleFavoriteResponse> = () =>
  new Promise((resolve) =>
    setTimeout(
      () =>
        resolve({ favorited: true, favorited_at: new Date().toISOString() }),
      300,
    ),
  );

const meta: Meta<typeof FileCard> = {
  title: "Dashboard/FileCard",
  component: FileCard,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Renders a file as a row (`list`) or tile (`grid`). Slots let callers mount a row menu (list only) and a favorite button. Click or Enter/Space fires `onOpen(file)` so the caller can open a viewer and fire `recordFileView`.",
      },
    },
  },
  argTypes: {
    variant: {
      control: { type: "radio" },
      options: ["list", "grid"],
    },
  },
  args: {
    onOpen: fn(),
  },
  decorators: [
    (Story, context) => (
      <div
        className={context.args.variant === "grid" ? "w-[220px]" : "w-[540px]"}
      >
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof FileCard>;

export const List: Story = {
  args: {
    file: makeFile(),
    variant: "list",
  },
};

export const ListWithFavorite: Story = {
  args: {
    file: makeFile({ favorited_at: new Date().toISOString() }),
    variant: "list",
    favoriteButton: (
      <FavoriteButton
        initialFavorited
        label="Favorite this file"
        size="sm"
        onToggle={mockFavoriteToggle}
      />
    ),
  },
};

export const ListWithRowMenu: Story = {
  args: {
    file: makeFile({
      last_viewed_at: new Date(Date.now() - 3 * 86_400_000).toISOString(),
    }),
    variant: "list",
    rowMenu: (
      <Button variant="ghost" size="sm" aria-label="File actions">
        <MoreVertical className="size-4" />
      </Button>
    ),
  },
};

export const ListPending: Story = {
  args: {
    file: makeFile({ status: "pending", size: 0 }),
    variant: "list",
  },
};

export const ListLongTitle: Story = {
  args: {
    file: makeFile({
      name: "A Very Long Filename That Should Truncate With An Ellipsis In The List Variant Without Ever Breaking The Row Height.pdf",
    }),
    variant: "list",
  },
};

export const ListImage: Story = {
  args: {
    file: makeFile({
      name: "diagram.png",
      mime_type: "image/png",
      size: 320_000,
    }),
    variant: "list",
  },
};

export const Grid: Story = {
  args: {
    file: makeFile(),
    variant: "grid",
  },
};

export const GridWithFavorite: Story = {
  args: {
    file: makeFile({ favorited_at: new Date().toISOString() }),
    variant: "grid",
    favoriteButton: (
      <FavoriteButton
        initialFavorited
        label="Favorite this file"
        size="sm"
        onToggle={mockFavoriteToggle}
      />
    ),
  },
};

export const GridPending: Story = {
  args: {
    file: makeFile({ status: "pending", size: 0 }),
    variant: "grid",
  },
};

export const UntitledFallback: Story = {
  args: {
    file: makeFile({ name: "" }),
    variant: "list",
  },
};
