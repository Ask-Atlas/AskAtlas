import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { RefSummary } from "@/lib/api/types";

import { CalloutBlock } from "./callout-block";
import { CourseRefCard } from "./course-ref-card";
import { EntityRefProvider } from "./entity-ref-context";
import type { EntityRef } from "./extract-refs";
import { FileRefCard } from "./file-ref-card";
import { QuizRefCard } from "./quiz-ref-card";
import { StudyGuideRefCard } from "./study-guide-ref-card";

const SG_ID = "11111111-2222-3333-4444-555555555555";
const QUIZ_ID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee";
const FILE_ID = "66666666-7777-8888-9999-000000000000";
const COURSE_ID = "cccccccc-dddd-eeee-ffff-111111111111";

const SAMPLE_REFS: Record<string, RefSummary | null> = {
  [`sg:${SG_ID}`]: {
    type: "sg",
    id: SG_ID,
    title: "Binary search trees: traversals + balancing",
    course: { department: "CPTS", number: "322" },
    quiz_count: 3,
    is_recommended: true,
  } as RefSummary,
  [`quiz:${QUIZ_ID}`]: {
    type: "quiz",
    id: QUIZ_ID,
    title: "BST practice set",
    question_count: 12,
    creator: { first_name: "Ada", last_name: "Lovelace" },
  } as RefSummary,
  [`file:${FILE_ID}`]: {
    type: "file",
    id: FILE_ID,
    name: "bst-cheatsheet.pdf",
    size: 184320,
    mime_type: "application/pdf",
    status: "complete",
  } as RefSummary,
  [`course:${COURSE_ID}`]: {
    type: "course",
    id: COURSE_ID,
    title: "Systems Programming",
    department: "CPTS",
    number: "322",
    school: { name: "Washington State University", acronym: "WSU" },
  } as RefSummary,
};

const REFS: EntityRef[] = [
  { type: "sg", id: SG_ID },
  { type: "quiz", id: QUIZ_ID },
  { type: "file", id: FILE_ID },
  { type: "course", id: COURSE_ID },
];

function AllCards({ inline = false }: { inline?: boolean }) {
  return (
    <div className="flex flex-col gap-2">
      {inline ? (
        <p className="text-sm leading-relaxed">
          Start with the <StudyGuideRefCard id={SG_ID} inline /> primer, take
          the <QuizRefCard id={QUIZ_ID} inline /> for{" "}
          <CourseRefCard id={COURSE_ID} inline />. Reference:{" "}
          <FileRefCard id={FILE_ID} inline />.
        </p>
      ) : (
        <>
          <StudyGuideRefCard id={SG_ID} />
          <QuizRefCard id={QUIZ_ID} />
          <FileRefCard id={FILE_ID} />
          <CourseRefCard id={COURSE_ID} />
        </>
      )}
    </div>
  );
}

const meta: Meta<typeof AllCards> = {
  title: "Dashboard/RefsCards",
  component: AllCards,
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "Direct renders of each entity ref card with a Provider seeded by `initial` -- bypasses the markdown pipeline so the card + context path can be verified independently.",
      },
    },
  },
  decorators: [
    (Story) => (
      <EntityRefProvider refs={REFS} initial={SAMPLE_REFS}>
        <div className="mx-auto max-w-2xl">
          <Story />
        </div>
      </EntityRefProvider>
    ),
  ],
};
export default meta;
type Story = StoryObj<typeof AllCards>;

export const LeafCardsPopulated: Story = { args: { inline: false } };
export const InlinePillsPopulated: Story = { args: { inline: true } };

export const LeafCardsMissing: Story = {
  decorators: [
    (Story) => (
      <EntityRefProvider
        refs={REFS}
        initial={{
          [`sg:${SG_ID}`]: null,
          [`quiz:${QUIZ_ID}`]: null,
          [`file:${FILE_ID}`]: null,
          [`course:${COURSE_ID}`]: null,
        }}
      >
        <div className="mx-auto max-w-2xl">
          <Story />
        </div>
      </EntityRefProvider>
    ),
  ],
  args: { inline: false },
};

export const CalloutVariants: Story = {
  render: () => (
    <div className="mx-auto flex max-w-2xl flex-col gap-3">
      <CalloutBlock type="note">
        Sequential consistency is strictly stronger than causal consistency.
      </CalloutBlock>
      <CalloutBlock type="warning">
        The textbook uses <em>linearizability</em> where some papers say{" "}
        <em>atomic consistency</em>.
      </CalloutBlock>
      <CalloutBlock type="tip">
        Reach for a mutex when exactly one writer should be in the critical
        section.
      </CalloutBlock>
    </div>
  ),
  decorators: [(Story) => <Story />],
};
