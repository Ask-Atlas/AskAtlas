import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { SkeletonGrid } from "./skeleton-grid";

const meta: Meta<typeof SkeletonGrid> = {
  title: "UI/SkeletonGrid",
  component: SkeletonGrid,
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Compound skeleton for card grids (file tiles, course tiles, quiz cards). Responsive 2 -> 3 -> 4 columns matches the real FileCard grid so the placeholder doesn't shift layout on data arrival.",
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="p-8">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof SkeletonGrid>;

export const Default: Story = {};

export const Small: Story = {
  args: { count: 3 },
};

export const Large: Story = {
  args: { count: 12 },
};
