import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { CalloutBlock } from "./callout-block";

describe("CalloutBlock", () => {
  it("renders as an aside with role=note", () => {
    render(<CalloutBlock type="note">body</CalloutBlock>);
    const el = screen.getByRole("note");
    expect(el.tagName).toBe("ASIDE");
    expect(el).toHaveAccessibleName("Note");
  });

  it("warning variant sets the Warning accessible name", () => {
    render(<CalloutBlock type="warning">watch out</CalloutBlock>);
    expect(screen.getByRole("note")).toHaveAccessibleName("Warning");
  });

  it("tip variant sets the Tip accessible name", () => {
    render(<CalloutBlock type="tip">pro tip</CalloutBlock>);
    expect(screen.getByRole("note")).toHaveAccessibleName("Tip");
  });

  it("unknown type falls back to note without crashing", () => {
    const warn = jest.spyOn(console, "warn").mockImplementation(() => {});
    render(<CalloutBlock type="danger">still renders</CalloutBlock>);
    expect(screen.getByRole("note")).toHaveAccessibleName("Note");
    expect(warn).toHaveBeenCalled();
    warn.mockRestore();
  });

  it("missing type defaults to note", () => {
    render(<CalloutBlock>body</CalloutBlock>);
    expect(screen.getByRole("note")).toHaveAccessibleName("Note");
  });

  it("renders children inside the body slot", () => {
    render(
      <CalloutBlock type="tip">
        <p>hello</p>
      </CalloutBlock>,
    );
    expect(screen.getByText("hello")).toBeInTheDocument();
  });
});
