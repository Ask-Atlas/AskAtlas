/**
 * Exercises the ASK-162 acceptance criteria for the grid variant:
 * default count renders 6 cards, explicit count wins, responsive grid
 * classes are present, and a custom className composes cleanly.
 */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { SkeletonGrid } from "./skeleton-grid";

describe("SkeletonGrid", () => {
  it("renders 6 cards by default", () => {
    render(<SkeletonGrid />);
    const grid = screen.getByTestId("skeleton-grid");
    expect(grid.children).toHaveLength(6);
  });

  it("renders the requested number of cards", () => {
    render(<SkeletonGrid count={4} />);
    const grid = screen.getByTestId("skeleton-grid");
    expect(grid.children).toHaveLength(4);
  });

  it("applies responsive grid columns", () => {
    render(<SkeletonGrid />);
    const grid = screen.getByTestId("skeleton-grid");
    // 2 -> 3 -> 4 columns keeps card grids legible on phone/tablet/desktop
    // without per-page overrides.
    expect(grid.className).toContain("grid-cols-2");
    expect(grid.className).toContain("md:grid-cols-3");
    expect(grid.className).toContain("lg:grid-cols-4");
  });

  it("forwards a custom className without dropping base layout", () => {
    render(<SkeletonGrid className="mt-8" />);
    const grid = screen.getByTestId("skeleton-grid");
    expect(grid.className).toContain("mt-8");
    expect(grid.className).toContain("grid");
  });

  it("is hidden from assistive tech", () => {
    render(<SkeletonGrid />);
    expect(screen.getByTestId("skeleton-grid")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });
});
