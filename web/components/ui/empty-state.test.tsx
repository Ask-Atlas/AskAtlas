/**
 * Exercises the ASK-161 acceptance criteria: title-only render hides the
 * optional slots, the full prop set renders icon/title/body/action in
 * DOM order, and a custom `className` merges via `cn()` without stomping
 * the base layout classes (so consumers can tweak spacing or width per
 * surface without reinventing the primitive).
 */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { EmptyState } from "./empty-state";

describe("EmptyState", () => {
  it("renders only the title when no optional props are provided", () => {
    render(<EmptyState title="No files yet" />);
    expect(
      screen.getByRole("heading", { name: "No files yet" }),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("empty-icon")).not.toBeInTheDocument();
    expect(screen.queryByTestId("empty-action")).not.toBeInTheDocument();
  });

  it("renders icon, title, body, and action in that DOM order", () => {
    const { container } = render(
      <EmptyState
        title="No files yet"
        body="Upload a file to get started."
        icon={<svg data-testid="empty-icon" />}
        action={
          <button type="button" data-testid="empty-action">
            Upload
          </button>
        }
      />,
    );
    const root = container.firstElementChild;
    expect(root).not.toBeNull();
    const children = Array.from(root?.children ?? []);
    // Slots must render in the documented visual order so consumers can
    // reason about layout without reading the implementation.
    expect(children[0]?.querySelector("[data-testid='empty-icon']")).not.toBe(
      null,
    );
    expect(children[1]?.tagName).toBe("H3");
    expect(children[2]?.tagName).toBe("P");
    expect(children[3]?.querySelector("[data-testid='empty-action']")).not.toBe(
      null,
    );
  });

  it("merges a custom className with the base layout classes", () => {
    const { container } = render(
      <EmptyState title="No files yet" className="border border-dashed" />,
    );
    const root = container.firstElementChild as HTMLElement;
    expect(root.className).toContain("border");
    // Base layout must still apply -- otherwise consumers could accidentally
    // break centering by passing any className.
    expect(root.className).toContain("flex");
    expect(root.className).toContain("text-center");
  });

  it("omits the body paragraph entirely when body is not passed", () => {
    const { container } = render(<EmptyState title="No files yet" />);
    expect(container.querySelector("p")).toBeNull();
  });
});
