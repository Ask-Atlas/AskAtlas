import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { QuizListItemResponse } from "@/lib/api/types";

import { QuizCard } from "./quiz-card";

function makeQuiz(
  overrides: Partial<QuizListItemResponse> = {},
): QuizListItemResponse {
  return {
    id: "q_preview_1",
    title: "CPTS 322 — Midterm Review",
    description: null,
    question_count: 12,
    creator: {
      id: "u_preview_1",
      first_name: "Ada",
      last_name: "Lovelace",
    },
    created_at: new Date(Date.now() - 6 * 86_400_000).toISOString(),
    updated_at: new Date(Date.now() - 2 * 86_400_000).toISOString(),
    ...overrides,
  };
}

const meta: Meta<typeof QuizCard> = {
  title: "Dashboard/QuizCard",
  component: QuizCard,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "List item for a quiz on the study-guide view page. Card tap opens the quiz detail; the inline Practice CTA routes straight into /practice?quiz={id}. Uses the stretched-link pattern so the Practice <Link> is a sibling of the overlay <Link>, not a descendant.",
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[620px]">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof QuizCard>;

export const Default: Story = {
  args: {
    quiz: makeQuiz(),
  },
};

export const SingleQuestion: Story = {
  args: {
    quiz: makeQuiz({ question_count: 1 }),
  },
};

export const ManyQuestions: Story = {
  args: {
    quiz: makeQuiz({ question_count: 42 }),
  },
};

export const LongTitle: Story = {
  args: {
    quiz: makeQuiz({
      title:
        "Comprehensive Midterm Review Covering Chapters 1 Through 7 With A Focus On Systems Programming And Concurrent Execution Models",
    }),
  },
};

export const CustomHref: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Demonstrates the `href` override, e.g. for a contextual view inside a study guide.",
      },
    },
  },
  args: {
    quiz: makeQuiz(),
    href: "/study-guides/g_preview_1/quizzes/q_preview_1",
  },
};
