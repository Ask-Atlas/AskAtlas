/**
 * Exercises the ASK-170 acceptance criteria: create-mode submit
 * with valid fields shapes the body correctly (AC1); edit-mode
 * pre-fills from `initial` (AC2); title/content below min-length
 * disables Save + shows inline error (AC3, AC4); Cancel callback
 * fires; the imperative `setError` hook surfaces server-side
 * validation errors for AC5. Also covers the tag-chip interactions
 * (Enter commits, dedupe, Backspace pops last chip, X removes).
 */
import { createRef } from "react";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

// ContentEditor pulls react-markdown (ESM) transitively. Tests here
// only care about the surrounding form behavior; stub to a plain
// textarea that preserves value/onChange/onBlur semantics.
jest.mock("./content-editor", () => {
  const React = jest.requireActual("react");
  return {
    ContentEditor: React.forwardRef(function StubContentEditor(
      props: {
        value: string;
        onChange: (v: string) => void;
        onBlur?: () => void;
        name?: string;
        placeholder?: string;
        rows?: number;
        disabled?: boolean;
      },
      ref: React.Ref<HTMLTextAreaElement>,
    ) {
      return (
        <textarea
          ref={ref}
          aria-label="Content"
          value={props.value}
          onChange={(e) => props.onChange(e.target.value)}
          onBlur={props.onBlur}
          name={props.name}
          placeholder={props.placeholder}
          rows={props.rows}
          disabled={props.disabled}
        />
      );
    }),
  };
});

import type { StudyGuideDetailResponse } from "@/lib/api/types";

import { StudyGuideForm, type StudyGuideFormHandle } from "./study-guide-form";

function makeStudyGuide(
  overrides: Partial<StudyGuideDetailResponse> = {},
): StudyGuideDetailResponse {
  return {
    id: "g_preview_1",
    title: "CPTS 322 Midterm Review",
    description: null,
    content: "# Chapter 1\n\nSystems programming is...",
    tags: ["midterm", "systems-programming"],
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
    created_at: "2026-04-20T10:00:00Z",
    updated_at: "2026-04-20T10:00:00Z",
    ...overrides,
  } as StudyGuideDetailResponse;
}

async function typeTag(user: ReturnType<typeof userEvent.setup>, tag: string) {
  // The tag input collapses to a "+ Add tag" button when idle -- open
  // it, type, submit, and the input keeps focus for rapid-add so the
  // next call re-queries the textbox afresh.
  const trigger = screen.queryByRole("button", { name: /add tag/i });
  if (trigger) {
    await user.click(trigger);
  }
  const input = screen.getByRole("textbox", { name: /new tag/i });
  await user.type(input, `${tag}{Enter}`);
}

describe("StudyGuideForm / create mode", () => {
  it("submits the shaped body when valid fields are filled (AC1)", async () => {
    const user = userEvent.setup();
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    render(
      <StudyGuideForm mode="create" onSubmit={onSubmit} onCancel={jest.fn()} />,
    );
    await user.type(screen.getByLabelText(/title/i), "Concurrency notes");
    await user.type(
      screen.getByLabelText(/content/i),
      "# Overview\n\nMutexes, semaphores, and monitors.",
    );
    await typeTag(user, "concurrency");
    await typeTag(user, "threads");
    await typeTag(user, "systems");
    await user.click(screen.getByRole("button", { name: /create/i }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      title: "Concurrency notes",
      content: "# Overview\n\nMutexes, semaphores, and monitors.",
      tags: ["concurrency", "threads", "systems"],
    });
  });

  it("disables Save and shows inline error while title is below 3 chars (AC3)", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await user.type(screen.getByLabelText(/title/i), "Hi");
    await user.type(
      screen.getByLabelText(/content/i),
      "Body long enough to satisfy the min-length requirement.",
    );
    expect(screen.getByRole("button", { name: /create/i })).toBeDisabled();
    await waitFor(() =>
      expect(
        screen.getByText("Title must be at least 3 characters"),
      ).toBeInTheDocument(),
    );
  });

  it("disables Save and shows inline error while content is below 10 chars (AC4)", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await user.type(screen.getByLabelText(/title/i), "Valid title");
    await user.type(screen.getByLabelText(/content/i), "short");
    expect(screen.getByRole("button", { name: /create/i })).toBeDisabled();
    await waitFor(() =>
      expect(
        screen.getByText("Content must be at least 10 characters"),
      ).toBeInTheDocument(),
    );
  });

  it("fires onCancel when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const onCancel = jest.fn();
    render(
      <StudyGuideForm mode="create" onSubmit={jest.fn()} onCancel={onCancel} />,
    );
    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});

describe("StudyGuideForm / tag chips", () => {
  it("adds a chip when the user types and presses Enter", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await typeTag(user, "midterm");
    const group = screen.getByRole("group", { name: /tags/i });
    expect(within(group).getByText("midterm")).toBeInTheDocument();
  });

  it("normalizes to lowercase + trims whitespace", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await typeTag(user, "  MidTERM  ");
    const group = screen.getByRole("group", { name: /tags/i });
    expect(within(group).getByText("midterm")).toBeInTheDocument();
    expect(within(group).queryByText("MidTERM")).not.toBeInTheDocument();
  });

  it("dedupes: re-adding an existing tag is a no-op and clears the input", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await typeTag(user, "midterm");
    await typeTag(user, "midterm");
    const group = screen.getByRole("group", { name: /tags/i });
    expect(within(group).getAllByText("midterm")).toHaveLength(1);
    const input = screen.getByRole("textbox", { name: /new tag/i });
    expect(input).toHaveValue("");
  });

  it("removes a chip when its X is clicked", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await typeTag(user, "midterm");
    await typeTag(user, "concurrency");
    await user.click(screen.getByRole("button", { name: /remove midterm/i }));
    const group = screen.getByRole("group", { name: /tags/i });
    expect(within(group).queryByText("midterm")).not.toBeInTheDocument();
    expect(within(group).getByText("concurrency")).toBeInTheDocument();
  });

  it("pops the last chip on Backspace when the draft input is empty", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await typeTag(user, "midterm");
    await typeTag(user, "concurrency");
    const input = screen.getByRole("textbox", { name: /new tag/i });
    input.focus();
    await user.keyboard("{Backspace}");
    const group = screen.getByRole("group", { name: /tags/i });
    expect(within(group).queryByText("concurrency")).not.toBeInTheDocument();
    expect(within(group).getByText("midterm")).toBeInTheDocument();
  });

  it("ignores empty / whitespace-only submits", async () => {
    const user = userEvent.setup();
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /add tags?/i }));
    const input = screen.getByRole("textbox", { name: /new tag/i });
    await user.type(input, "   {Enter}");
    const group = screen.getByRole("group", { name: /tags/i });
    // No chip was added -- the only button in the group is the
    // reopened "+ Add tag" trigger.
    expect(within(group).queryAllByRole("button", { name: /remove/i })).toHaveLength(0);
  });
});

describe("StudyGuideForm / edit mode", () => {
  it("pre-fills the form from `initial` (AC2)", () => {
    const initial = makeStudyGuide();
    render(
      <StudyGuideForm
        mode="edit"
        initial={initial}
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    expect(screen.getByLabelText(/title/i)).toHaveValue(
      "CPTS 322 Midterm Review",
    );
    expect(screen.getByLabelText(/content/i)).toHaveValue(
      "# Chapter 1\n\nSystems programming is...",
    );
    const group = screen.getByRole("group", { name: /tags/i });
    expect(within(group).getByText("midterm")).toBeInTheDocument();
    expect(within(group).getByText("systems-programming")).toBeInTheDocument();
  });

  it("labels the submit button 'Save' in edit mode", () => {
    render(
      <StudyGuideForm
        mode="edit"
        initial={makeStudyGuide()}
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: /^save$/i })).toBeInTheDocument();
  });

  it("round-trips pre-filled tags back through submit unchanged", async () => {
    const user = userEvent.setup();
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    render(
      <StudyGuideForm
        mode="edit"
        initial={makeStudyGuide()}
        onSubmit={onSubmit}
        onCancel={jest.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^save$/i }));
    await waitFor(() =>
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          tags: ["midterm", "systems-programming"],
        }),
      ),
    );
  });
});

describe("StudyGuideForm / server error surface (AC5)", () => {
  it("exposes a ref.setError that shows a field-level message", async () => {
    const ref = createRef<StudyGuideFormHandle>();
    render(
      <StudyGuideForm
        ref={ref}
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    // Simulates a caller projecting ApiError.body.details onto fields:
    //   for (const d of err.body.details) ref.setError(d.field, d.message);
    await act(async () => {
      ref.current?.setError("title", "Title already taken");
    });
    await waitFor(() =>
      expect(screen.getByText("Title already taken")).toBeInTheDocument(),
    );
  });
});
