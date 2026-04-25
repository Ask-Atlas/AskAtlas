import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { StudyGuideDetailResponse } from "@/lib/api/types";

import type { GrantsManagerActions } from "./grants-manager";
import { StudyGuideForm } from "./study-guide-form";

// Storybook can't reach the real server actions (Vite would pull in
// `@clerk/nextjs/server` which is Node-only). Provide a stub bag so
// the GrantsManager mounts and lands in its "No shares yet" state.
const stubGrantActions: GrantsManagerActions = {
  listGrants: async () => ({ grants: [] }),
  listEnrollments: async () => ({ enrollments: [] }),
  createGrant: async (studyGuideId, body) => ({
    id: "preview-grant",
    study_guide_id: studyGuideId,
    grantee_type: body.grantee_type,
    grantee_id: body.grantee_id,
    permission: body.permission,
    granted_by: "u_preview_1",
    created_at: "2026-04-01T00:00:00Z",
  }),
  revokeGrant: async () => {},
};

function makeStudyGuide(
  overrides: Partial<StudyGuideDetailResponse> = {},
): StudyGuideDetailResponse {
  return {
    id: "g_preview_1",
    title: "CPTS 322 Midterm Review",
    description: null,
    content:
      "# Overview\n\nThis guide covers mutexes, semaphores, and monitors in the context of the CPTS 322 midterm.\n\n## Mutex vs Semaphore\n\nA mutex is owned by a single thread; a semaphore counts permits.",
    tags: ["midterm", "concurrency", "systems-programming"],
    creator: { id: "u_preview_1", first_name: "Ada", last_name: "Lovelace" },
    course: {
      id: "c_preview_1",
      department: "CPTS",
      number: "322",
      title: "Systems Programming",
    },
    vote_score: 0,
    user_vote: null,
    view_count: 0,
    is_recommended: false,
    recommended_by: [],
    quizzes: [],
    resources: [],
    files: [],
    created_at: "2026-04-17T00:00:00Z",
    updated_at: "2026-04-22T00:00:00Z",
    ...overrides,
  } as StudyGuideDetailResponse;
}

function resolveAfter(ms: number): () => Promise<void> {
  return () => new Promise((resolve) => setTimeout(resolve, ms));
}

const meta: Meta<typeof StudyGuideForm> = {
  title: "Dashboard/StudyGuideForm",
  component: StudyGuideForm,
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "Shared create + edit form backed by react-hook-form + Zod + shadcn Form primitives. Save is reactively disabled until title (≥3) and content (≥10) pass validation. Caller handles success redirect + error toast; server-side validation errors are surfaced by projecting ApiError details onto fields via the exposed `setError` ref.",
      },
    },
  },
  argTypes: {
    mode: {
      control: { type: "radio" },
      options: ["create", "edit"],
    },
  },
  decorators: [
    (Story) => (
      <div className="mx-auto max-w-2xl">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof StudyGuideForm>;

export const CreateEmpty: Story = {
  args: {
    mode: "create",
    onSubmit: resolveAfter(500),
    onCancel: () => {},
  },
};

export const EditPrefilled: Story = {
  args: {
    mode: "edit",
    initial: makeStudyGuide(),
    onSubmit: resolveAfter(500),
    onCancel: () => {},
  },
};

export const SlowSubmit: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Simulates a 2s round-trip so you can see the button label swap to 'Creating…' and Cancel disable.",
      },
    },
  },
  args: {
    mode: "create",
    onSubmit: resolveAfter(2000),
    onCancel: () => {},
  },
};

export const EditEmptyTags: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Edit mode when the guide has no tags -- verifies the tags input pre-fills with an empty string.",
      },
    },
  },
  args: {
    mode: "edit",
    initial: makeStudyGuide({ tags: [] }),
    onSubmit: resolveAfter(500),
    onCancel: () => {},
  },
};

export const CreatePublic: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Create mode with the visibility chip pre-flipped to Public. Clicking the chip opens the popover showing the Private/Public segmented control plus a hint that grants require saving the guide first.",
      },
    },
  },
  args: {
    mode: "create",
    initial: makeStudyGuide({ visibility: "public" }),
    onSubmit: resolveAfter(500),
    onCancel: () => {},
  },
};

export const EditWithGrantsStub: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "Edit mode variant for the visibility popover. Click the chip to see the Private/Public toggle plus the GrantsManager surface. Storybook injects stub actions so the manager lands in its 'No shares yet' state -- real grant flows are covered by the Jest/RTL tests.",
      },
    },
  },
  args: {
    mode: "edit",
    initial: makeStudyGuide({ visibility: "private" }),
    onSubmit: resolveAfter(500),
    onCancel: () => {},
    grantActions: stubGrantActions,
  },
};
