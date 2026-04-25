/**
 * ASK-197 acceptance tests for the course detail page.
 *
 * Server Components return a Promise<ReactElement>; we await the page
 * function directly and feed the resolved JSX into RTL `render`. The
 * `@/lib/api` actions and `next/navigation` `notFound` are mocked at
 * the module level so the test owns every API response and the 404
 * branch can assert without needing the real error boundary.
 */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { ApiError } from "@/lib/api/errors";
import type {
  CourseDetailResponse,
  ListMyEnrollmentsResponse,
  ListStudyGuidesResponse,
} from "@/lib/api/types";

import CourseDetailPage from "./page";

jest.mock("../../../../lib/api", () => ({
  getCourse: jest.fn(),
  listCourseStudyGuides: jest.fn(),
  listMyEnrollments: jest.fn(),
  joinSection: jest.fn(),
  leaveSection: jest.fn(),
}));

const notFoundMock = jest.fn(() => {
  throw new Error("NEXT_NOT_FOUND");
});

jest.mock("next/navigation", () => ({
  notFound: () => notFoundMock(),
}));

import {
  getCourse,
  listCourseStudyGuides,
  listMyEnrollments,
} from "../../../../lib/api";

const getCourseMock = getCourse as jest.MockedFunction<typeof getCourse>;
const listGuidesMock = listCourseStudyGuides as jest.MockedFunction<
  typeof listCourseStudyGuides
>;
const listEnrollmentsMock = listMyEnrollments as jest.MockedFunction<
  typeof listMyEnrollments
>;

const SECTION_A = "11111111-1111-1111-1111-111111111111";
const SECTION_B = "22222222-2222-2222-2222-222222222222";

function makeCourse(
  overrides: Partial<CourseDetailResponse> = {},
): CourseDetailResponse {
  return {
    id: "c_preview_1",
    school: {
      id: "s_preview_1",
      name: "Washington State University",
      acronym: "WSU",
      city: "Pullman",
      state: "WA",
      country: "US",
    },
    department: "CPTS",
    number: "322",
    title: "Systems Programming",
    description: null,
    created_at: "2026-04-20T10:00:00Z",
    sections: [
      {
        id: SECTION_A,
        term: "Spring 2026",
        section_code: "01",
        instructor_name: "Dr. Aragoneses",
        member_count: 47,
      },
      {
        id: SECTION_B,
        term: "Spring 2026",
        section_code: "02",
        instructor_name: "Dr. Williams",
        member_count: 38,
      },
    ],
    ...overrides,
  };
}

function makeGuides(count: number = 0): ListStudyGuidesResponse {
  return {
    study_guides: Array.from({ length: count }).map((_, i) => ({
      id: `g_${i}`,
      title: `Guide ${i + 1}`,
      description: null,
      tags: [],
      creator: {
        id: `u_${i}`,
        first_name: "Sarah",
        last_name: "K.",
      },
      course_id: "c_preview_1",
      vote_score: 10,
      view_count: 0,
      is_recommended: false,
      quiz_count: 0,
      visibility: "public",
      created_at: "2026-04-20T10:00:00Z",
      updated_at: "2026-04-20T10:00:00Z",
    })),
    next_cursor: null,
    has_more: false,
  };
}

function makeEnrollments(sectionId?: string): ListMyEnrollmentsResponse {
  if (!sectionId) return { enrollments: [] };
  return {
    enrollments: [
      {
        section: {
          id: sectionId,
          term: "Spring 2026",
          section_code: "01",
          instructor_name: "Dr. Aragoneses",
        },
        course: {
          id: "c_preview_1",
          department: "CPTS",
          number: "322",
          title: "Systems Programming",
        },
        school: { id: "s_preview_1", acronym: "WSU" },
        role: "student",
        joined_at: "2026-04-20T10:00:00Z",
      },
    ],
  };
}

beforeEach(() => {
  jest.clearAllMocks();
});

async function renderPage() {
  const ui = await CourseDetailPage({
    params: Promise.resolve({ courseId: "c_preview_1" }),
  });
  render(ui);
}

describe("CourseDetailPage (ASK-197)", () => {
  it("renders course header with title, school, and metadata (AC1)", async () => {
    getCourseMock.mockResolvedValue(makeCourse());
    listGuidesMock.mockResolvedValue(makeGuides(2));
    listEnrollmentsMock.mockResolvedValue(makeEnrollments());

    await renderPage();

    expect(
      screen.getByRole("heading", { level: 1, name: "Systems Programming" }),
    ).toBeInTheDocument();
    expect(screen.getByText("CPTS 322")).toBeInTheDocument();
    expect(screen.getByText("Washington State University")).toBeInTheDocument();
    expect(screen.getByText(/85 enrolled/)).toBeInTheDocument();
    expect(screen.getByText(/2 sections/)).toBeInTheDocument();
  });

  it("shows section picker with Join controls when not enrolled (AC2 not-enrolled half)", async () => {
    getCourseMock.mockResolvedValue(makeCourse());
    listGuidesMock.mockResolvedValue(makeGuides(2));
    listEnrollmentsMock.mockResolvedValue(makeEnrollments());

    await renderPage();

    expect(
      screen.getByRole("heading", { name: "Pick a section" }),
    ).toBeInTheDocument();
    const joinButtons = screen.getAllByRole("button", { name: "Join" });
    expect(joinButtons).toHaveLength(2);
  });

  it("renders enrolled banner + Enrolled control for the joined section (AC2)", async () => {
    getCourseMock.mockResolvedValue(makeCourse());
    listGuidesMock.mockResolvedValue(makeGuides(2));
    listEnrollmentsMock.mockResolvedValue(makeEnrollments(SECTION_A));

    await renderPage();

    expect(screen.getByText(/You.*re in/i)).toBeInTheDocument();
    expect(screen.getByText("Section 01")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /Study guides/ }),
    ).toBeInTheDocument();
  });

  it("renders 'No study guides yet' empty state when enrolled and zero guides (AC3)", async () => {
    getCourseMock.mockResolvedValue(makeCourse());
    listGuidesMock.mockResolvedValue(makeGuides(0));
    listEnrollmentsMock.mockResolvedValue(makeEnrollments(SECTION_A));

    await renderPage();

    expect(screen.getByText("No study guides yet")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Create one/i })).toHaveAttribute(
      "href",
      "/study-guides/new?course=c_preview_1",
    );
  });

  it("renders 'No sections offered' when course has zero sections (AC4)", async () => {
    getCourseMock.mockResolvedValue(makeCourse({ sections: [] }));
    listGuidesMock.mockResolvedValue(makeGuides(0));
    listEnrollmentsMock.mockResolvedValue(makeEnrollments());

    await renderPage();

    expect(screen.getByText("No sections offered")).toBeInTheDocument();
  });

  it("triggers notFound() when getCourse returns 404 (AC5)", async () => {
    const apiErr = new ApiError(
      "GET /courses/c_preview_1 failed: 404",
      { status: 404 } as unknown as Response,
      null,
    );
    getCourseMock.mockRejectedValue(apiErr);
    listGuidesMock.mockResolvedValue(makeGuides(0));
    listEnrollmentsMock.mockResolvedValue(makeEnrollments());

    await expect(renderPage()).rejects.toThrow("NEXT_NOT_FOUND");
    expect(notFoundMock).toHaveBeenCalledTimes(1);
  });
});
