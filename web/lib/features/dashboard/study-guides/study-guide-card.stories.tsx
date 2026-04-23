import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { StudyGuideListItemResponse } from "./study-guide-card";
import { StudyGuideCard } from "./study-guide-card";

function makeGuide(
  overrides: Partial<StudyGuideListItemResponse> = {},
): StudyGuideListItemResponse {
  return {
    id: "sg_preview_1",
    title: "Linear Algebra — Eigenvalues & Eigenvectors Cheatsheet",
    creator: { display_name: "Aiko Tanaka" },
    vote_score: 42,
    quiz_count: 4,
    is_recommended: true,
    tags: ["week 6", "midterm", "eigen"],
    course_id: "c_math340",
    ...overrides,
  };
}

const meta: Meta<typeof StudyGuideCard> = {
  title: "Dashboard/StudyGuideCard",
  component: StudyGuideCard,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Renders a community study guide either as a full `list` card (title, creator, course, vote + quiz badges, tags) or a `compact` card (title + author + quiz count). Wraps the whole thing in a Next `<Link>` to the guide detail page.",
      },
    },
  },
  argTypes: {
    variant: {
      control: { type: "radio" },
      options: ["list", "compact"],
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[360px]">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof StudyGuideCard>;

export const List: Story = {
  args: {
    guide: makeGuide(),
    variant: "list",
    courseLabel: "MATH 340",
  },
};

export const ListWithoutRecommendedBadge: Story = {
  args: {
    guide: makeGuide({ is_recommended: false, vote_score: 3 }),
    variant: "list",
    courseLabel: "MATH 340",
  },
};

export const ListNoTags: Story = {
  args: {
    guide: makeGuide({ tags: [], vote_score: 0 }),
    variant: "list",
    courseLabel: "MATH 340",
  },
};

export const ListNoCourseLabel: Story = {
  args: {
    guide: makeGuide(),
    variant: "list",
  },
};

export const SingleQuiz: Story = {
  args: {
    guide: makeGuide({ quiz_count: 1, tags: ["week 1"] }),
    variant: "list",
    courseLabel: "MATH 340",
  },
};

export const Compact: Story = {
  args: {
    guide: makeGuide(),
    variant: "compact",
  },
};

export const CompactSingleQuiz: Story = {
  args: {
    guide: makeGuide({ quiz_count: 1, title: "Quick review before the quiz" }),
    variant: "compact",
  },
};

export const LongTitleOverflow: Story = {
  args: {
    guide: makeGuide({
      title:
        "A Very Long Study Guide Title That Should Line-Clamp To Two Lines So The Card Layout Never Breaks No Matter What",
    }),
    variant: "list",
    courseLabel: "MATH 340",
  },
};
