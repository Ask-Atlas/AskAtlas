/**
 * Acceptance tests for the "My Study Guides" page.
 *
 * Server Components return a Promise<ReactElement>; we await the page
 * function directly and feed the resolved JSX into RTL `render`. The
 * `@/lib/api` server actions are mocked at the module level so the
 * test owns every API response.
 */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import type { ListMyStudyGuidesResponse } from "@/lib/api/types";

import MyStudyGuidesPage from "./page";

jest.mock("../../../../lib/api", () => ({
  listMyStudyGuides: jest.fn(),
}));

import { listMyStudyGuides } from "../../../../lib/api";

const listMyStudyGuidesMock = listMyStudyGuides as jest.MockedFunction<
  typeof listMyStudyGuides
>;

type MyGuide = ListMyStudyGuidesResponse["study_guides"][number];

function makeGuide(overrides: Partial<MyGuide> = {}): MyGuide {
  return {
    id: "g_1",
    title: "Linear Algebra Cheat Sheet",
    description: "Eigenvalues, determinants, and the SVD.",
    tags: ["math"],
    creator: { id: "u_1", first_name: "Ada", last_name: "Lovelace" },
    course_id: "c_1",
    vote_score: 12,
    view_count: 0,
    is_recommended: false,
    quiz_count: 3,
    created_at: "2026-04-20T10:00:00Z",
    updated_at: "2026-04-20T10:00:00Z",
    deleted_at: null,
    visibility: "private",
    ...overrides,
  };
}

function makeResponse(guides: MyGuide[] = []): ListMyStudyGuidesResponse {
  return { study_guides: guides, next_cursor: null, has_more: false };
}

beforeEach(() => {
  jest.clearAllMocks();
});

async function renderPage() {
  const ui = await MyStudyGuidesPage();
  render(ui);
}

describe("MyStudyGuidesPage", () => {
  it("renders the empty state when the caller has no guides", async () => {
    listMyStudyGuidesMock.mockResolvedValue(makeResponse([]));

    await renderPage();

    expect(
      screen.getByRole("heading", { level: 1, name: "My study guides" }),
    ).toBeInTheDocument();
    expect(screen.getByText("No study guides yet")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /create one/i })).toHaveAttribute(
      "href",
      "/study-guides/new",
    );
  });

  it("renders a card for each live guide", async () => {
    listMyStudyGuidesMock.mockResolvedValue(
      makeResponse([
        makeGuide({ id: "g_a", title: "Algebra Notes" }),
        makeGuide({ id: "g_b", title: "Calculus Outline" }),
      ]),
    );

    await renderPage();

    expect(screen.getByText("Algebra Notes")).toBeInTheDocument();
    expect(screen.getByText("Calculus Outline")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("hides soft-deleted guides from the live list", async () => {
    listMyStudyGuidesMock.mockResolvedValue(
      makeResponse([
        makeGuide({ id: "g_live", title: "Live Guide" }),
        makeGuide({
          id: "g_dead",
          title: "Deleted Guide",
          deleted_at: "2026-04-21T10:00:00Z",
        }),
      ]),
    );

    await renderPage();

    expect(screen.getByText("Live Guide")).toBeInTheDocument();
    expect(screen.queryByText("Deleted Guide")).not.toBeInTheDocument();
  });
});
