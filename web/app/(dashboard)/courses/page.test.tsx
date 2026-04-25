import { act, fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import type {
  CourseResponse,
  ListCoursesQuery,
  ListCoursesResponse,
} from "../../../lib/api/types";

import CoursesPage from "./page";

jest.mock("../../../lib/api", () => ({
  listCourses: jest.fn(),
  listSchools: jest.fn(),
  listMyEnrollments: jest.fn(),
}));

jest.mock("../../../lib/features/shared/toast/toast", () => ({
  toast: { error: jest.fn(), success: jest.fn(), info: jest.fn() },
}));

import { listCourses, listMyEnrollments, listSchools } from "../../../lib/api";

const listCoursesMock = listCourses as jest.MockedFunction<typeof listCourses>;
const listSchoolsMock = listSchools as jest.MockedFunction<typeof listSchools>;
const listEnrollmentsMock = listMyEnrollments as jest.MockedFunction<
  typeof listMyEnrollments
>;

function makeCourse(overrides: Partial<CourseResponse> = {}): CourseResponse {
  return {
    id: `c-${Math.random().toString(36).slice(2, 8)}`,
    school: {
      id: "s-1",
      name: "Atlas University",
      acronym: "AU",
      city: "Pullman",
      state: "WA",
      country: "US",
    },
    department: "CPTS",
    number: "322",
    title: "Systems Programming",
    description: null,
    created_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function makeResponse(
  courses: CourseResponse[],
  next_cursor: string | null = null,
): ListCoursesResponse {
  return {
    courses,
    next_cursor,
    has_more: next_cursor !== null,
  };
}

async function flushPromises() {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
}

describe("CoursesPage (ASK-193)", () => {
  beforeEach(() => {
    listSchoolsMock.mockResolvedValue({
      schools: [],
      next_cursor: null,
      has_more: false,
    });
    listEnrollmentsMock.mockResolvedValue({ enrollments: [] });
    listCoursesMock.mockResolvedValue(makeResponse([]));
  });

  afterEach(() => {
    jest.clearAllMocks();
    jest.useRealTimers();
  });

  it("renders course cards once the first page returns (AC1)", async () => {
    listCoursesMock.mockResolvedValueOnce(
      makeResponse([
        makeCourse({
          id: "c1",
          department: "CPTS",
          number: "322",
          title: "Systems Programming",
        }),
        makeCourse({
          id: "c2",
          department: "MATH",
          number: "216",
          title: "Discrete Mathematics",
        }),
      ]),
    );

    render(<CoursesPage />);
    await flushPromises();

    expect(screen.getByText("CPTS 322")).toBeInTheDocument();
    expect(screen.getByText("Systems Programming")).toBeInTheDocument();
    expect(screen.getByText("MATH 216")).toBeInTheDocument();
    expect(listCoursesMock).toHaveBeenCalledWith({});
  });

  it("refetches with q after the SearchInput debounce (AC2)", async () => {
    jest.useFakeTimers();
    listCoursesMock.mockResolvedValue(makeResponse([]));

    render(<CoursesPage />);
    await act(async () => {
      await Promise.resolve();
    });

    const search = screen.getByLabelText("Search courses") as HTMLInputElement;
    fireEvent.change(search, { target: { value: "algo" } });

    act(() => {
      jest.advanceTimersByTime(250);
    });
    await act(async () => {
      await Promise.resolve();
    });

    const queries = listCoursesMock.mock.calls.map(
      (call) => call[0] as ListCoursesQuery,
    );
    expect(queries.some((q) => q?.q === "algo")).toBe(true);
  });

  it("appends results when Load more is clicked (AC3)", async () => {
    listCoursesMock.mockResolvedValueOnce(
      makeResponse(
        [makeCourse({ id: "p1", number: "100", title: "Intro Page 1" })],
        "cursor-1",
      ),
    );
    listCoursesMock.mockResolvedValueOnce(
      makeResponse(
        [makeCourse({ id: "p2", number: "200", title: "Intro Page 2" })],
        null,
      ),
    );

    render(<CoursesPage />);
    await flushPromises();

    expect(screen.getByText("Intro Page 1")).toBeInTheDocument();

    const loadMore = screen.getByRole("button", { name: /load more/i });
    fireEvent.click(loadMore);
    await flushPromises();

    expect(screen.getByText("Intro Page 1")).toBeInTheDocument();
    expect(screen.getByText("Intro Page 2")).toBeInTheDocument();
    expect(listCoursesMock).toHaveBeenLastCalledWith({ cursor: "cursor-1" });
  });

  it("renders an EmptyState with Clear filters when zero results (AC4)", async () => {
    listCoursesMock.mockResolvedValue(makeResponse([]));

    render(<CoursesPage />);
    await flushPromises();

    expect(screen.getByText("No courses match")).toBeInTheDocument();
    const clear = screen.getByRole("button", { name: /clear filters/i });
    expect(clear).toBeInTheDocument();
    fireEvent.click(clear);
    await flushPromises();
    expect(listCoursesMock).toHaveBeenLastCalledWith({});
  });
});
