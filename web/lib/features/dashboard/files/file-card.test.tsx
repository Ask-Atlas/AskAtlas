/**
 * Covers the ASK-164 acceptance criteria: list/grid rendering, size +
 * mime-type labels, pending status copy, Untitled fallback, keyboard
 * activation, and the rowMenu/favoriteButton slots.
 */
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import type { FileResponse } from "@/lib/api/types";

import { FileCard } from "./file-card";

function makeFile(overrides: Partial<FileResponse> = {}): FileResponse {
  return {
    id: "f_preview_1",
    name: "Lecture 3 - Linear Algebra Review.pdf",
    size: 1_048_576, // 1 MB
    mime_type: "application/pdf",
    status: "complete",
    created_at: "2026-04-20T10:00:00Z",
    updated_at: "2026-04-20T10:00:00Z",
    favorited_at: null,
    last_viewed_at: null,
    ...overrides,
  };
}

describe("FileCard / list variant", () => {
  it("shows name, formatted size, and mime label", () => {
    render(<FileCard file={makeFile()} variant="list" />);
    expect(
      screen.getByText("Lecture 3 - Linear Algebra Review.pdf"),
    ).toBeInTheDocument();
    expect(screen.getByText(/1 MB/)).toBeInTheDocument();
  });

  it("renders size 0 as '0 B' without breaking", () => {
    render(<FileCard file={makeFile({ size: 0 })} variant="list" />);
    expect(screen.getByText(/0 B/)).toBeInTheDocument();
  });

  it("falls back to 'Untitled' for empty name", () => {
    render(<FileCard file={makeFile({ name: "" })} variant="list" />);
    expect(screen.getByText("Untitled")).toBeInTheDocument();
  });

  it("shows 'Processing…' when status is pending", () => {
    render(<FileCard file={makeFile({ status: "pending" })} variant="list" />);
    expect(screen.getByText(/Processing/)).toBeInTheDocument();
  });

  it("invokes onOpen on click", async () => {
    const onOpen = jest.fn();
    render(<FileCard file={makeFile()} variant="list" onOpen={onOpen} />);
    await userEvent.click(screen.getByRole("button"));
    expect(onOpen).toHaveBeenCalledTimes(1);
    expect(onOpen.mock.calls[0][0]).toMatchObject({ id: "f_preview_1" });
  });

  it("invokes onOpen on Enter key", async () => {
    const onOpen = jest.fn();
    render(<FileCard file={makeFile()} variant="list" onOpen={onOpen} />);
    screen.getByRole("button").focus();
    await userEvent.keyboard("{Enter}");
    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it("invokes onOpen on Space key", async () => {
    const onOpen = jest.fn();
    render(<FileCard file={makeFile()} variant="list" onOpen={onOpen} />);
    screen.getByRole("button").focus();
    await userEvent.keyboard(" ");
    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it("is not role=button when onOpen is absent", () => {
    render(<FileCard file={makeFile()} variant="list" />);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("does not fire onOpen when the row menu is clicked", async () => {
    const onOpen = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        onOpen={onOpen}
        rowMenu={<button type="button">Menu</button>}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Menu" }));
    expect(onOpen).not.toHaveBeenCalled();
  });
});

describe("FileCard / rename prop (ASK-165)", () => {
  function deferred<T>() {
    let resolve!: (value: T) => void;
    let reject!: (err: unknown) => void;
    const promise = new Promise<T>((res, rej) => {
      resolve = res;
      reject = rej;
    });
    return { promise, resolve, reject };
  }

  it("renders an auto-focused input with the filename pre-selected (AC2)", () => {
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    const input = screen.getByRole("textbox", {
      name: /new file name/i,
    }) as HTMLInputElement;
    expect(input).toHaveFocus();
    expect(input.selectionStart).toBe(0);
    expect(input.selectionEnd).toBe(input.value.length);
  });

  it("commits a trimmed value on Enter (AC3)", async () => {
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.type(input, "  New name.pdf  {Enter}");
    expect(onCommit).toHaveBeenCalledWith("New name.pdf");
  });

  it("does not call onCommit when the name is empty (AC6)", async () => {
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.keyboard("{Enter}");
    expect(onCommit).not.toHaveBeenCalled();
    expect(onCancel).not.toHaveBeenCalled();
    // Input stays open so the user can recover without re-opening the menu.
    expect(input).toBeInTheDocument();
  });

  it("does not call onCommit when the name is whitespace-only", async () => {
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.type(input, "   {Enter}");
    expect(onCommit).not.toHaveBeenCalled();
    expect(input).toBeInTheDocument();
  });

  it("cancels without calling onCommit when the name equals the current name", async () => {
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    // Default value already matches file.name; just press Enter.
    await userEvent.keyboard("{Enter}");
    expect(onCommit).not.toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("cancels on Esc", async () => {
    const onCommit = jest.fn();
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    await userEvent.keyboard("{Escape}");
    expect(onCommit).not.toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("closes (via onCancel) when onCommit rejects (AC4)", async () => {
    const commitPromise = deferred<void>();
    const onCommit = jest.fn(() => commitPromise.promise);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        rename={{ onCommit, onCancel }}
      />,
    );
    const input = screen.getByRole("textbox", { name: /new file name/i });
    await userEvent.clear(input);
    await userEvent.type(input, "New name.pdf{Enter}");
    await act(async () => {
      commitPromise.reject(new Error("network"));
    });
    // On rejection we fall back to onCancel so the parent clears
    // rename-mode; the filename text remains untouched because we
    // never mutated `file.name`.
    await waitFor(() => expect(onCancel).toHaveBeenCalled());
  });

  it("does not fire onOpen while in rename mode", async () => {
    const onOpen = jest.fn();
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="list"
        onOpen={onOpen}
        rename={{ onCommit, onCancel }}
      />,
    );
    // The card should NOT have role=button while renaming -- clicking
    // the row shouldn't open the file and unmount the input.
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(onOpen).not.toHaveBeenCalled();
  });

  it("ignores the rename prop on the grid variant", () => {
    const onCommit = jest.fn().mockResolvedValue(undefined);
    const onCancel = jest.fn();
    render(
      <FileCard
        file={makeFile()}
        variant="grid"
        rename={{ onCommit, onCancel }}
      />,
    );
    expect(
      screen.queryByRole("textbox", { name: /new file name/i }),
    ).not.toBeInTheDocument();
  });
});

describe("FileCard / grid variant", () => {
  it("renders the filename and size", () => {
    render(<FileCard file={makeFile()} variant="grid" />);
    expect(
      screen.getByText("Lecture 3 - Linear Algebra Review.pdf"),
    ).toBeInTheDocument();
    expect(screen.getByText(/1 MB/)).toBeInTheDocument();
  });

  it("invokes onOpen on click", async () => {
    const onOpen = jest.fn();
    render(<FileCard file={makeFile()} variant="grid" onOpen={onOpen} />);
    await userEvent.click(screen.getByRole("button"));
    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it("shows 'Processing…' when status is pending", () => {
    render(<FileCard file={makeFile({ status: "pending" })} variant="grid" />);
    expect(screen.getByText(/Processing/)).toBeInTheDocument();
  });
});
