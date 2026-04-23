/**
 * Exercises the ConfirmationDialog contract: it's visible when `open`,
 * invokes `onConfirm` on confirm click (may return a promise the caller
 * awaits), does NOT auto-close on confirm (caller owns close), fires
 * `onOpenChange(false)` when Cancel closes it, the destructive flag
 * swaps the confirm button variant, and `disabled` locks both buttons.
 */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import { ConfirmationDialog } from "./confirmation-dialog";

function renderDialog(
  overrides: Partial<Parameters<typeof ConfirmationDialog>[0]> = {},
) {
  const props = {
    open: true,
    onOpenChange: jest.fn(),
    title: "Delete file",
    description: "This can't be undone.",
    onConfirm: jest.fn(),
    ...overrides,
  };
  return { props, ...render(<ConfirmationDialog {...props} />) };
}

describe("ConfirmationDialog", () => {
  it("renders the title and description when open", () => {
    renderDialog();
    expect(
      screen.getByRole("alertdialog", { name: /delete file/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("This can't be undone.")).toBeInTheDocument();
  });

  it("does not render content when closed", () => {
    renderDialog({ open: false });
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("invokes onConfirm when the confirm button is clicked", async () => {
    const { props } = renderDialog({ confirmLabel: "Delete" });
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(props.onConfirm).toHaveBeenCalledTimes(1);
  });

  it("stays open during async confirm so callers can await the promise", async () => {
    let released = false;
    const onConfirm = jest.fn(
      () =>
        new Promise<void>((resolve) => {
          setTimeout(() => {
            released = true;
            resolve();
          }, 10);
        }),
    );
    const { props } = renderDialog({ onConfirm, confirmLabel: "Delete" });
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    // Critical: the primitive must NOT have closed the dialog itself. If
    // downstream tickets await an API call before closing, a self-closing
    // dialog would unmount the spinner and re-enable buttons prematurely.
    expect(props.onOpenChange).not.toHaveBeenCalledWith(false);
    await onConfirm.mock.results[0]?.value;
    expect(released).toBe(true);
  });

  it("fires onOpenChange(false) when Cancel is clicked", async () => {
    const { props } = renderDialog({ cancelLabel: "Nevermind" });
    await userEvent.click(screen.getByRole("button", { name: "Nevermind" }));
    expect(props.onOpenChange).toHaveBeenCalledWith(false);
  });

  it("applies destructive styling when `destructive` is true", () => {
    renderDialog({ destructive: true, confirmLabel: "Delete" });
    expect(screen.getByRole("button", { name: "Delete" })).toHaveClass(
      "bg-destructive",
    );
  });

  it("uses the default variant when `destructive` is false", () => {
    renderDialog({ confirmLabel: "Save" });
    expect(screen.getByRole("button", { name: "Save" })).not.toHaveClass(
      "bg-destructive",
    );
  });

  it("disables both buttons when `disabled` is true", async () => {
    const { props } = renderDialog({
      disabled: true,
      confirmLabel: "Delete",
      cancelLabel: "Nevermind",
    });
    const confirm = screen.getByRole("button", { name: "Delete" });
    const cancel = screen.getByRole("button", { name: "Nevermind" });
    expect(confirm).toBeDisabled();
    expect(cancel).toBeDisabled();

    await userEvent.click(confirm);
    expect(props.onConfirm).not.toHaveBeenCalled();
  });
});
