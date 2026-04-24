import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { RefSummary } from "@/lib/api/types";

import { ArticleRenderer } from "./article-renderer";

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

const meta: Meta<typeof ArticleRenderer> = {
  title: "Dashboard/ArticleRenderer",
  component: ArticleRenderer,
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "Renders study-guide markdown as a styled article. GFM (tables, task lists, strikethrough) + inline HTML via rehype-raw behind rehype-sanitize, Next.js Link for internal hrefs, external hrefs open in a new tab with rel=noopener noreferrer. Pair with tailwindcss-typography's `prose` classes for typography.",
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="mx-auto max-w-3xl">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof ArticleRenderer>;

export const BasicMarkdown: Story = {
  args: {
    content: [
      "# CPTS 322 Midterm Review",
      "",
      "## Mutex vs Semaphore",
      "",
      "A **mutex** is owned by a single thread; a *semaphore* counts permits.",
      "",
      "- Mutexes have an owner",
      "- Semaphores count",
      "- Monitors wrap both with a condition variable",
      "",
      "> Rule of thumb: reach for a mutex when only one writer should be in the critical section.",
      "",
      "```c",
      "pthread_mutex_lock(&m);",
      "// critical section",
      "pthread_mutex_unlock(&m);",
      "```",
    ].join("\n"),
  },
};

export const GfmTable: Story = {
  args: {
    content: [
      "## Primitive comparison",
      "",
      "| Primitive | Owned by | Counts | Blocks writers |",
      "| --------- | -------- | ------ | -------------- |",
      "| Mutex     | Yes      | No     | Yes            |",
      "| Semaphore | No       | Yes    | Optional       |",
      "| Monitor   | Yes      | No     | Yes            |",
    ].join("\n"),
  },
};

export const TaskList: Story = {
  args: {
    content: [
      "## Midterm checklist",
      "",
      "- [x] Review lock-free queues",
      "- [x] Re-derive Peterson's algorithm",
      "- [ ] Bench semaphores vs monitors",
      "- [ ] Skim the deadlock-detection chapter",
    ].join("\n"),
  },
};

export const InternalAndExternalLinks: Story = {
  args: {
    content: [
      "See the [BST study guide](/study-guides/abc-123) for a primer, or jump straight to [practice](/practice?quiz=xyz).",
      "",
      "External reference: [OS principles (arXiv)](https://arxiv.org/abs/1234).",
    ].join("\n"),
  },
};

export const WithImage: Story = {
  args: {
    content: [
      "## Visualising the happens-before relation",
      "",
      "![Happens-before lattice with message passing arrows](https://picsum.photos/seed/happens-before/960/540)",
      "",
      "The lattice above shows event ordering between processes.",
    ].join("\n"),
  },
};

export const BrokenImage: Story = {
  args: {
    content: [
      "## When a referenced file is missing",
      "",
      "![A chart that no longer exists](/api/files/deleted-file-id/download)",
      "",
      "The chart above was attached to this guide originally but has since been removed.",
    ].join("\n"),
  },
};

export const CalloutAside: Story = {
  parameters: {
    docs: {
      description: {
        story:
          'Inline `<aside class="callout">` HTML is allowed via rehype-raw and survives rehype-sanitize. Useful for author-flagged asides without needing a custom directive.',
      },
    },
  },
  args: {
    content: [
      "## Consistency",
      "",
      "Sequential consistency is strictly stronger than causal consistency.",
      "",
      '<aside class="callout rounded-md border-l-4 border-amber-400 bg-amber-50 p-4 text-amber-900">',
      "  <strong>Heads up:</strong> the textbook uses `linearizability`",
      "  where some papers still say `atomic consistency`.",
      "</aside>",
    ].join("\n"),
  },
};

export const XssAttemptStripped: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Inline <script> tags, javascript: hrefs, and other unsafe primitives are stripped before render.",
      },
    },
  },
  args: {
    content: [
      "## Sanitized content",
      "",
      "The block below contains a `<script>` tag and a `javascript:` link; both should be gone by the time you read this.",
      "",
      "<script>window.__xss = true;</script>",
      "",
      "[do not click me](javascript:alert('xss'))",
      "",
      "The paragraph text still renders.",
    ].join("\n"),
  },
};

export const EmptyContent: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Whitespace-only content renders nothing -- callers get a tidy no-op rather than a stray wrapper div.",
      },
    },
  },
  args: { content: "   " },
};

export const EntityRefLeafCards: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Leaf-position directives render as block-level cards. Uses `initialRefs` so the preview does not hit the live API.",
      },
    },
  },
  args: {
    initialRefs: SAMPLE_REFS,
    content: [
      "## See also",
      "",
      `::sg{id="${SG_ID}"}`,
      "",
      `::quiz{id="${QUIZ_ID}"}`,
      "",
      `::file{id="${FILE_ID}"}`,
      "",
      `::course{id="${COURSE_ID}"}`,
    ].join("\n"),
  },
};

export const EntityRefInlinePills: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Inline directives (single colon) render as compact pills alongside prose.",
      },
    },
  },
  args: {
    initialRefs: SAMPLE_REFS,
    content: `Start with the :sg{id="${SG_ID}"} primer, then take the :quiz{id="${QUIZ_ID}"} for :course{id="${COURSE_ID}"}. Reference: :file{id="${FILE_ID}"}.`,
  },
};

export const MissingRefFallback: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "When a ref resolves to `null` (deleted or invisible), the card renders a muted Unavailable placeholder instead of crashing.",
      },
    },
  },
  args: {
    initialRefs: { [`sg:${SG_ID}`]: null },
    content: `This study guide was removed: ::sg{id="${SG_ID}"}.`,
  },
};

export const Callouts: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Container directives map to styled callout blocks. Three variants: note / warning / tip. Unknown types fall back to note.",
      },
    },
  },
  args: {
    initialRefs: {},
    content: [
      ':::callout{type="note"}',
      "Sequential consistency is strictly stronger than causal consistency.",
      ":::",
      "",
      ':::callout{type="warning"}',
      "The textbook uses *linearizability* where some papers say *atomic consistency*.",
      ":::",
      "",
      ':::callout{type="tip"}',
      "Reach for a mutex when exactly one writer should be in the critical section.",
      ":::",
    ].join("\n"),
  },
};

export const UnknownDirectivesPassThrough: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Unknown directive names (e.g. `::foo{}`) and directives with invalid UUIDs render as plain text -- no crash, no card.",
      },
    },
  },
  args: {
    initialRefs: {},
    content: [
      '`::foo{id="xyz"}` renders as plain text:',
      "",
      '::foo{id="xyz"}',
      "",
      'And `::sg{id="not-a-uuid"}` also passes through:',
      "",
      '::sg{id="not-a-uuid"}',
    ].join("\n"),
  },
};
