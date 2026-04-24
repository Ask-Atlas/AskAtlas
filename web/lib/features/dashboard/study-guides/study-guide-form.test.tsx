/**
 * Exercises the ASK-170 acceptance criteria: create-mode submit
 * with valid fields shapes the body correctly (AC1); edit-mode
 * pre-fills from `initial` (AC2); title/content below min-length
 * disables Save + shows inline error (AC3, AC4); Cancel callback
 * fires; tags are comma-split/trimmed on submit; and the imperative
 * `setError` hook surfaces server-side validation errors for AC5.
 */
import { createRef } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

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

describe("StudyGuideForm / create mode", () => {
  it("submits the shaped body when valid fields are filled (AC1)", async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    render(
      <StudyGuideForm mode="create" onSubmit={onSubmit} onCancel={jest.fn()} />,
    );
    await userEvent.type(screen.getByLabelText(/title/i), "Concurrency notes");
    await userEvent.type(
      screen.getByLabelText(/content/i),
      "# Overview\n\nMutexes, semaphores, and monitors.",
    );
    await userEvent.type(
      screen.getByLabelText(/tags/i),
      "concurrency, threads, systems",
    );
    await userEvent.click(screen.getByRole("button", { name: /create/i }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      title: "Concurrency notes",
      content: "# Overview\n\nMutexes, semaphores, and monitors.",
      tags: ["concurrency", "threads", "systems"],
    });
  });

  it("trims whitespace around comma-separated tags", async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    render(
      <StudyGuideForm mode="create" onSubmit={onSubmit} onCancel={jest.fn()} />,
    );
    await userEvent.type(screen.getByLabelText(/title/i), "Title here");
    await userEvent.type(
      screen.getByLabelText(/content/i),
      "Body long enough to pass validation threshold.",
    );
    await userEvent.type(
      screen.getByLabelText(/tags/i),
      "  midterm ,  concurrency  ,,  systems  ",
    );
    await userEvent.click(screen.getByRole("button", { name: /create/i }));
    await waitFor(() =>
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          tags: ["midterm", "concurrency", "systems"],
        }),
      ),
    );
  });

  it("disables Save and shows inline error while title is below 3 chars (AC3)", async () => {
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await userEvent.type(screen.getByLabelText(/title/i), "Hi");
    await userEvent.type(
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
    render(
      <StudyGuideForm
        mode="create"
        onSubmit={jest.fn()}
        onCancel={jest.fn()}
      />,
    );
    await userEvent.type(screen.getByLabelText(/title/i), "Valid title");
    await userEvent.type(screen.getByLabelText(/content/i), "short");
    expect(screen.getByRole("button", { name: /create/i })).toBeDisabled();
    await waitFor(() =>
      expect(
        screen.getByText("Content must be at least 10 characters"),
      ).toBeInTheDocument(),
    );
  });

  it("fires onCancel when Cancel is clicked", async () => {
    const onCancel = jest.fn();
    render(
      <StudyGuideForm mode="create" onSubmit={jest.fn()} onCancel={onCancel} />,
    );
    await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onCancel).toHaveBeenCalledTimes(1);
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
    expect(screen.getByLabelText(/tags/i)).toHaveValue(
      "midterm, systems-programming",
    );
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
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    render(
      <StudyGuideForm
        mode="edit"
        initial={makeStudyGuide()}
        onSubmit={onSubmit}
        onCancel={jest.fn()}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));
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
