/**
 * Exercises the ASK-165 acceptance criteria: menu opens with Rename +
 * Delete, Rename opens the inline input auto-focused and pre-selected,
 * Enter saves when changed, empty stays open, identical-name closes
 * without an API call, rejection reverts + closes, Delete opens the
 * confirmation dialog and confirming fires onDelete.
 */
import { act, render, screen, waitFor } from "@testing-library/react";
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

describe("FileRowMenu / menu trigger", () => {
  it("shows Rename and Delete when the trigger is clicked (AC1)", async () => {
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={jest.fn().mockResolvedValue(undefined)}
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

describe("FileRowMenu / rename flow", () => {
  async function openRename() {
    await userEvent.click(
      screen.getByRole("button", { name: /file actions/i }),
    );
    await userEvent.click(screen.getByRole("menuitem", { name: "Rename" }));
  }

  it("focuses the input with the current filename pre-selected (AC2)", async () => {
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={jest.fn().mockResolvedValue(undefined)}
        onDelete={jest.fn()}
      />,
    );
    await openRename();
    const input = screen.getByRole("textbox", {
      name: /new file name/i,
    }) as HTMLInputElement;
    expect(input).toHaveFocus();
    expect(input.selectionStart).toBe(0);
    expect(input.selectionEnd).toBe(input.value.length);
  });

  it("calls onRename with the trimmed value when Enter is pressed (AC3)", async () => {
    const onRename = jest.fn().mockResolvedValue(undefined);
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={onRename}
        onDelete={jest.fn()}
      />,
    );
    await openRename();
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.type(input, "  New name.pdf  {Enter}");
    expect(onRename).toHaveBeenCalledWith("New name.pdf");
  });

  it("does not call onRename when the new name is empty (AC6)", async () => {
    const onRename = jest.fn().mockResolvedValue(undefined);
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={onRename}
        onDelete={jest.fn()}
      />,
    );
    await openRename();
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.keyboard("{Enter}");
    expect(onRename).not.toHaveBeenCalled();
    // Input must stay open so the user can recover without re-opening
    // the dropdown.
    expect(input).toBeInTheDocument();
  });

  it("does not call onRename when the new name is whitespace-only", async () => {
    const onRename = jest.fn().mockResolvedValue(undefined);
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={onRename}
        onDelete={jest.fn()}
      />,
    );
    await openRename();
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.type(input, "   {Enter}");
    expect(onRename).not.toHaveBeenCalled();
    expect(input).toBeInTheDocument();
  });

  it("does not call onRename when the new name equals the current name (edge case)", async () => {
    const onRename = jest.fn().mockResolvedValue(undefined);
    const file = makeFile();
    render(
      <FileRowMenu file={file} onRename={onRename} onDelete={jest.fn()} />,
    );
    await openRename();
    // Default value already matches file.name; just press Enter.
    await userEvent.keyboard("{Enter}");
    expect(onRename).not.toHaveBeenCalled();
    // And the input closes because the user's intent was "no change".
    await waitFor(() =>
      expect(
        screen.queryByRole("textbox", { name: /new file name/i }),
      ).not.toBeInTheDocument(),
    );
  });

  it("cancels on Esc and closes the input without calling onRename", async () => {
    const onRename = jest.fn();
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={onRename}
        onDelete={jest.fn()}
      />,
    );
    await openRename();
    await userEvent.keyboard("{Escape}");
    expect(onRename).not.toHaveBeenCalled();
    await waitFor(() =>
      expect(
        screen.queryByRole("textbox", { name: /new file name/i }),
      ).not.toBeInTheDocument(),
    );
  });

  it("reverts to the menu trigger when onRename rejects (AC4)", async () => {
    const renamePromise = deferred<void>();
    const onRename = jest.fn(() => renamePromise.promise);
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={onRename}
        onDelete={jest.fn()}
      />,
    );
    await openRename();
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.type(input, "New name.pdf{Enter}");
    await act(async () => {
      renamePromise.reject(new Error("network"));
    });
    // After the rejection settles the menu returns -- caller toasts and
    // the row re-renders with the original filename (which is untouched
    // since we never mutated `file`).
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: /file actions/i }),
      ).toBeInTheDocument(),
    );
  });
});

describe("FileRowMenu / delete flow", () => {
  it("opens the confirmation dialog without firing onDelete", async () => {
    const onDelete = jest.fn().mockResolvedValue(undefined);
    render(
      <FileRowMenu
        file={makeFile()}
        onRename={jest.fn()}
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
        onRename={jest.fn()}
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
        onRename={jest.fn()}
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
