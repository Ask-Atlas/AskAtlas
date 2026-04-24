/**
 * ASK-206 unit coverage. Tests the pieces we own directly (internal-
 * href detection + ArticleImage loading/loaded/error states). The
 * end-to-end markdown pipeline -- GFM tables, task lists, rehype-raw
 * passthrough, rehype-sanitize XSS stripping, Next Link routing --
 * lives in Storybook stories + E2E, because react-markdown and the
 * unified ecosystem ship ESM-only and jest's default node_modules
 * ignore rule eats them.
 */
import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { ArticleImage, isInternalHref } from "./article-internals";

describe("isInternalHref", () => {
  it.each([
    ["/", true],
    ["/study-guides/abc", true],
    ["/practice?quiz=xyz", true],
    ["/files/abc/download", true],
  ])("treats %s as internal", (href, expected) => {
    expect(isInternalHref(href)).toBe(expected);
  });

  it.each([
    ["https://example.com", false],
    ["http://example.com", false],
    ["//evil.example/steal", false], // protocol-relative is NOT internal
    ["mailto:a@b.com", false],
    ["javascript:alert(1)", false],
    ["#section", false],
    ["./relative", false],
  ])("treats %s as external", (href, expected) => {
    expect(isInternalHref(href)).toBe(expected);
  });
});

describe("ArticleImage", () => {
  it("renders nothing when src is missing", () => {
    const { container } = render(<ArticleImage alt="x" />);
    expect(container.firstChild).toBeNull();
  });

  it("shows the skeleton while loading and hides the <img>", () => {
    render(<ArticleImage src="/api/files/abc/download" alt="diagram" />);
    expect(screen.getByTestId("article-image-skeleton")).toBeInTheDocument();
    const img = screen.getByAltText("diagram");
    expect(img).toHaveClass("hidden");
  });

  it("renders the alt as a figcaption when provided", () => {
    render(<ArticleImage src="/api/files/abc/download" alt="diagram caption" />);
    const caption = screen.getByText("diagram caption");
    expect(caption.tagName).toBe("FIGCAPTION");
  });

  it("omits the figcaption when alt is empty", () => {
    const { container } = render(<ArticleImage src="/api/files/abc/download" />);
    expect(container.querySelector("figcaption")).toBeNull();
  });

  it("swaps skeleton for the image once it loads", () => {
    render(<ArticleImage src="/api/files/abc/download" alt="diagram" />);
    const img = screen.getByAltText("diagram");
    fireEvent.load(img);
    expect(
      screen.queryByTestId("article-image-skeleton"),
    ).not.toBeInTheDocument();
    expect(img).toHaveClass("block");
  });

  it("renders the error fallback on image error (AC5)", () => {
    render(<ArticleImage src="/api/files/missing/download" alt="broken" />);
    const img = screen.getByAltText("broken");
    fireEvent.error(img);
    expect(screen.getByRole("alert")).toHaveTextContent("Image failed to load");
  });

  it("uses alt='' for unlabeled images (accessibility hygiene)", () => {
    // alt="" gives the img presentation role, so getByRole("img") skips
    // it; query the DOM directly.
    const { container } = render(<ArticleImage src="/api/files/abc/download" />);
    const img = container.querySelector("img");
    expect(img).toHaveAttribute("alt", "");
  });
});
