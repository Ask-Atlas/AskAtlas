import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { SkeletonList } from "./skeleton-list";

const meta: Meta<typeof SkeletonList> = {
  title: "UI/SkeletonList",
  component: SkeletonList,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Compound skeleton for row-based lists (files, study guides, resources). Callers render it while the data query is pending and swap it for the real list when it resolves.",
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[540px] rounded-md border">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof SkeletonList>;

export const Default: Story = {};

export const Single: Story = {
  args: { count: 1 },
};

export const LongList: Story = {
  args: { count: 8 },
};
