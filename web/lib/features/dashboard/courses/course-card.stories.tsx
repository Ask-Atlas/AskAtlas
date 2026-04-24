import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { CourseResponse } from "@/lib/api/types";

import { CourseCard } from "./course-card";

function makeCourse(overrides: Partial<CourseResponse> = {}): CourseResponse {
  return {
    id: "c_preview_1",
    school: {
      id: "s_preview_1",
      name: "Washington State University",
      acronym: "WSU",
      city: "Pullman",
      state: "WA",
      country: "US",
    },
    department: "CPTS",
    number: "322",
    title: "Systems Programming",
    description: null,
    created_at: "2026-04-20T10:00:00Z",
    ...overrides,
  };
}

const meta: Meta<typeof CourseCard> = {
  title: "Dashboard/CourseCard",
  component: CourseCard,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Card for a single course. Row variant is compact (catalog list, sidebar). Tile variant surfaces the school name for the dashboard tile grid and /me/courses. The outer element is a Next.js Link; rightSlot clicks don't navigate so interactive affordances (Join / Leave / Favorite) work unambiguously.",
      },
    },
  },
  argTypes: {
    variant: {
      control: { type: "radio" },
      options: ["row", "tile"],
    },
  },
  decorators: [
    (Story, context) => (
      <div
        className={context.args.variant === "tile" ? "w-[260px]" : "w-[540px]"}
      >
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof CourseCard>;

export const Row: Story = {
  args: {
    course: makeCourse(),
    variant: "row",
  },
};

export const RowWithJoinedBadge: Story = {
  args: {
    course: makeCourse(),
    variant: "row",
    rightSlot: (
      <Badge
        variant="secondary"
        className="bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400"
      >
        Joined
      </Badge>
    ),
  },
};

export const RowWithJoinCta: Story = {
  args: {
    course: makeCourse(),
    variant: "row",
    rightSlot: (
      <Button size="sm" variant="outline">
        Join
      </Button>
    ),
  },
};

export const RowLongTitle: Story = {
  args: {
    course: makeCourse({
      title:
        "Advanced Topics in Computer Architecture and Operating System Design",
    }),
    variant: "row",
  },
};

export const Tile: Story = {
  args: {
    course: makeCourse(),
    variant: "tile",
  },
};

export const TileWithJoinedBadge: Story = {
  args: {
    course: makeCourse(),
    variant: "tile",
    rightSlot: (
      <Badge
        variant="secondary"
        className="bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400"
      >
        Joined
      </Badge>
    ),
  },
};

export const TileLongTitle: Story = {
  args: {
    course: makeCourse({
      title:
        "Advanced Topics in Computer Architecture and Operating System Design",
    }),
    variant: "tile",
  },
};

export const TileAltSchool: Story = {
  args: {
    course: makeCourse({
      school: {
        id: "s_preview_2",
        name: "University of California, Berkeley",
        acronym: "UCB",
        city: "Berkeley",
        state: "CA",
        country: "US",
      },
      department: "CS",
      number: "162",
      title: "Operating Systems and System Programming",
    }),
    variant: "tile",
  },
};
