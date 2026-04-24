/**
 * Exercises the ASK-165 acceptance criteria for the menu shell:
 * trigger opens Rename + Delete, Rename fires `onStartRename` (the
 * rename UI itself lives on FileCard via its `rename` prop, tested
 * separately), Delete opens the shared ConfirmationDialog, confirm
 * fires onDelete, cancel does not.
 */
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import type { FileResponse } from "@/lib/api/types";

import { FileRowMenu } from "./file-row-menu";

function makeFile(overrides: Partial<FileResponse> = {}): FileResponse {
  return {
    id: "f_preview_1",
    name: "Lecture 3 - Linear Algebra Review.pdf",
    size: 1_048_576,
    mime_type: "application/pdf",
    status: "complete",
    created_at: "2026-04-20T10:00:00Z",
    updated_at: "2026-04-20T10:00:00Z",
    favorited_at: null,
    last_viewed_at: null,
    ...overrides,
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (err: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe("FileRowMenu / menu trigger (AC1)", () => {
  it("shows Rename and Delete when the trigger is clicked", async () => {
    render(
      <FileRowMenu
        file={makeFile()}
        onStartRename={jest.fn()}
        onDelete={jest.fn().mockResolvedValue(undefined)}
      />,
    );
    await userEvent.click(
      screen.getByRole("button", { name: /file actions/i }),
    );
    expect(screen.getByRole("menuitem", { name: "Rename" })).toBeVisible();
    expect(screen.getByRole("menuitem", { name: "Delete" })).toBeVisible();
  });
});

describe("FileRowMenu / rename intent", () => {
  it("fires onStartRename when the Rename item is picked", async () => {
    const onStartRename = jest.fn();
    render(
      <FileRowMenu
        file={makeFile()}
        onStartRename={onStartRename}
        onDelete={jest.fn()}
      />,
    );
    await userEvent.click(
      screen.getByRole("button", { name: /file actions/i }),
    );
    await userEvent.click(screen.getByRole("menuitem", { name: "Rename" }));
    // Inputs, commits, and cancels are owned by FileCard's `rename`
    // prop -- the menu's job is purely to signal intent.
    expect(onStartRename).toHaveBeenCalledTimes(1);
  });
});

describe("FileRowMenu / delete flow", () => {
  it("opens the confirmation dialog without firing onDelete", async () => {
    const onDelete = jest.fn().mockResolvedValue(undefined);
    render(
      <FileRowMenu
        file={makeFile()}
        onStartRename={jest.fn()}
        onDelete={onDelete}
      />,
    );
    await userEvent.click(
      screen.getByRole("button", { name: /file actions/i }),
    );
    await userEvent.click(screen.getByRole("menuitem", { name: "Delete" }));
    expect(
      screen.getByRole("alertdialog", { name: /delete file/i }),
    ).toBeInTheDocument();
    expect(onDelete).not.toHaveBeenCalled();
  });

  it("fires onDelete when the dialog is confirmed (AC5)", async () => {
    const deletePromise = deferred<void>();
    const onDelete = jest.fn(() => deletePromise.promise);
    render(
      <FileRowMenu
        file={makeFile()}
        onStartRename={jest.fn()}
        onDelete={onDelete}
      />,
    );
    await userEvent.click(
      screen.getByRole("button", { name: /file actions/i }),
    );
    await userEvent.click(screen.getByRole("menuitem", { name: "Delete" }));
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(onDelete).toHaveBeenCalledTimes(1);
    // Settle the transition so React doesn't warn about unfinished
    // action scope when the test unmounts.
    await act(async () => {
      deletePromise.resolve();
    });
  });

  it("does not fire onDelete when the dialog is cancelled", async () => {
    const onDelete = jest.fn().mockResolvedValue(undefined);
    render(
      <FileRowMenu
        file={makeFile()}
        onStartRename={jest.fn()}
        onDelete={onDelete}
      />,
    );
    await userEvent.click(
      screen.getByRole("button", { name: /file actions/i }),
    );
    await userEvent.click(screen.getByRole("menuitem", { name: "Delete" }));
    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onDelete).not.toHaveBeenCalled();
  });
});
