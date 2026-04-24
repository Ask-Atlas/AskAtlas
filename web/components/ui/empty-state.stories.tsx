import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { FileText, Inbox, Search } from "lucide-react";

import { Button } from "@/components/ui/button";

import { EmptyState } from "./empty-state";

const meta: Meta<typeof EmptyState> = {
  title: "UI/EmptyState",
  component: EmptyState,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Tier A primitive for list/grid surfaces that have nothing to render. Slots (icon, body, action) are all optional so the same component covers everything from a tiny inline 'no results' message to a full zero-state with illustration and CTA.",
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[420px] rounded-md border border-dashed">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof EmptyState>;

export const TitleOnly: Story = {
  args: {
    title: "No files yet",
  },
};

export const WithBody: Story = {
  args: {
    title: "No files yet",
    body: "Upload a PDF, slide deck, or image to get started.",
  },
};

export const WithIcon: Story = {
  args: {
    title: "No files yet",
    body: "Upload a PDF, slide deck, or image to get started.",
    icon: <Inbox className="size-10" />,
  },
};

export const FullZeroState: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Everything: icon, title, body, CTA. Canonical shape for a list page with no rows.",
      },
    },
  },
  args: {
    title: "No files yet",
    body: "Upload a PDF, slide deck, or image to get started.",
    icon: <FileText className="size-10" />,
    action: <Button>Upload a file</Button>,
  },
};

export const NoSearchResults: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Common variant for filtered lists when a query returns nothing.",
      },
    },
  },
  args: {
    title: "No results",
    body: 'Nothing matches "linear algebra". Try a different search.',
    icon: <Search className="size-10" />,
  },
};
