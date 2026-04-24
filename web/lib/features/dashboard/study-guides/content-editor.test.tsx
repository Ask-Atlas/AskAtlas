import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

// ArticleRenderer pulls react-markdown (ESM) transitively. Preview
// tab is not the subject of these tests -- stub to a simple <pre>.
jest.mock("./article-renderer", () => ({
  ArticleRenderer: ({ content }: { content: string }) => (
    <pre data-testid="preview-stub">{content}</pre>
  ),
}));

import { ContentEditor } from "./content-editor";

const SG_ID = "11111111-2222-3333-4444-555555555555";
const HOSTS = ["askatlas.app"];

function Harness({ initial = "" }: { initial?: string }) {
  const [v, setV] = useState(initial);
  return <ContentEditor value={v} onChange={setV} allowedHosts={HOSTS} />;
}

function pasteInto(textarea: HTMLTextAreaElement, text: string) {
  fireEvent.paste(textarea, {
    clipboardData: {
      getData: (t: string) => (t === "text/plain" ? text : ""),
    },
  });
}

describe("ContentEditor paste rewrite", () => {
  it("replaces a pasted app URL with the matching directive", () => {
    render(<Harness />);
    const ta = screen.getByRole("textbox") as HTMLTextAreaElement;
    ta.focus();
    pasteInto(ta, `https://askatlas.app/study-guides/${SG_ID}`);
    expect(ta.value).toBe(`::sg{id="${SG_ID}"}`);
  });

  it("preserves surrounding text and inserts at the caret", () => {
    render(<Harness initial="before  after" />);
    const ta = screen.getByRole("textbox") as HTMLTextAreaElement;
    ta.setSelectionRange("before ".length, "before ".length);
    pasteInto(ta, `https://askatlas.app/study-guides/${SG_ID}`);
    expect(ta.value).toBe(`before ::sg{id="${SG_ID}"} after`);
  });

  it("falls through to default paste for non-app URLs", () => {
    render(<Harness initial="" />);
    const ta = screen.getByRole("textbox") as HTMLTextAreaElement;
    pasteInto(ta, `https://evil.example/study-guides/${SG_ID}`);
    expect(ta.value).toBe("");
  });

  it("falls through to default paste when clipboard has multiple URLs", () => {
    render(<Harness initial="" />);
    const ta = screen.getByRole("textbox") as HTMLTextAreaElement;
    pasteInto(
      ta,
      `https://askatlas.app/study-guides/${SG_ID} https://askatlas.app/study-guides/${SG_ID}`,
    );
    expect(ta.value).toBe("");
  });

  it("falls through for unknown app routes", () => {
    render(<Harness />);
    const ta = screen.getByRole("textbox") as HTMLTextAreaElement;
    pasteInto(ta, `https://askatlas.app/users/${SG_ID}`);
    expect(ta.value).toBe("");
  });
});

describe("ContentEditor tabs", () => {
  it("renders Write and Preview tabs", () => {
    render(<Harness />);
    expect(screen.getByRole("tab", { name: /write/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /preview/i })).toBeInTheDocument();
  });

  it("switching to Preview shows the ArticleRenderer stub with current content", async () => {
    const user = userEvent.setup();
    render(<Harness initial="# Hello" />);
    await user.click(screen.getByRole("tab", { name: /preview/i }));
    expect(screen.getByTestId("preview-stub")).toHaveTextContent("# Hello");
  });

  it("empty content shows a 'Nothing to preview yet.' placeholder", async () => {
    const user = userEvent.setup();
    render(<Harness initial="  " />);
    await user.click(screen.getByRole("tab", { name: /preview/i }));
    expect(screen.getByText(/nothing to preview yet/i)).toBeInTheDocument();
    expect(screen.queryByTestId("preview-stub")).not.toBeInTheDocument();
  });

  it("switching back to Write keeps the in-progress value", async () => {
    const user = userEvent.setup();
    render(<Harness initial="in progress" />);
    await user.click(screen.getByRole("tab", { name: /preview/i }));
    await user.click(screen.getByRole("tab", { name: /write/i }));
    const ta = screen.getByRole("textbox") as HTMLTextAreaElement;
    expect(ta.value).toBe("in progress");
  });
});
