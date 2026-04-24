/**
 * Exercises the ASK-162 acceptance criteria for the list variant:
 * the default count renders 3 rows, an explicit count wins, and the
 * container receives a custom className without losing the base layout.
 */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { SkeletonList } from "./skeleton-list";

describe("SkeletonList", () => {
  it("renders 3 rows by default", () => {
    render(<SkeletonList />);
    const list = screen.getByTestId("skeleton-list");
    expect(list.children).toHaveLength(3);
  });

  it("renders the requested number of rows", () => {
    render(<SkeletonList count={5} />);
    const list = screen.getByTestId("skeleton-list");
    expect(list.children).toHaveLength(5);
  });

  it("forwards a custom className without dropping base layout", () => {
    render(<SkeletonList className="max-w-md" />);
    const list = screen.getByTestId("skeleton-list");
    expect(list.className).toContain("max-w-md");
    expect(list.className).toContain("flex");
  });

  it("is hidden from assistive tech", () => {
    render(<SkeletonList />);
    // Callers pair this with their own aria-live region (e.g. 'Loading
    // files…') so screen readers get one semantic announcement instead
    // of N pulsing placeholders.
    expect(screen.getByTestId("skeleton-list")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });
});
