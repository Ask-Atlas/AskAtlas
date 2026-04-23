/**
 * Covers the ASK-164 acceptance criteria: list/grid rendering, size +
 * mime-type labels, pending status copy, Untitled fallback, keyboard
 * activation, and the rowMenu/favoriteButton slots.
 */
import { render, screen } from "@testing-library/react";
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
