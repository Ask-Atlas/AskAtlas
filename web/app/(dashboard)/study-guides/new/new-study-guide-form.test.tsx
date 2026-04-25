/**
 * ASK-191 acceptance tests for the create-a-study-guide page.
 * Mocks the API actions, the toast, and `next/navigation` so the
 * test owns the redirect + error surface.
 */
import * as React from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import { ApiError } from "@/lib/api/errors";
import type { EnrollmentResponse } from "@/lib/api/types";

// ContentEditor transitively imports react-markdown (ESM); the same
// stub the StudyGuideForm test uses keeps Jest from choking on it.
jest.mock(
  "../../../../lib/features/dashboard/study-guides/content-editor",
  () => ({
    ContentEditor: React.forwardRef<
      unknown,
      { value: string; onChange: (next: string) => void }
    >(function StubContentEditor({ value, onChange }) {
      return (
        <textarea
          aria-label="Content"
          value={value}
          onChange={(event) => onChange(event.target.value)}
        />
      );
    }),
  }),
);

import { NewStudyGuideForm } from "./new-study-guide-form";

const pushMock = jest.fn();
const backMock = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
    back: backMock,
    refresh: jest.fn(),
    replace: jest.fn(),
    forward: jest.fn(),
    prefetch: jest.fn(),
  }),
}));

jest.mock("../../../../lib/api", () => ({
  createStudyGuideForCourse: jest.fn(),
}));

const toastErrorMock = jest.fn();
jest.mock("../../../../lib/features/shared/toast/toast", () => ({
  toast: {
    error: (err: unknown) => toastErrorMock(err),
    success: jest.fn(),
    info: jest.fn(),
  },
}));

import { createStudyGuideForCourse } from "../../../../lib/api";
const createMock = createStudyGuideForCourse as jest.MockedFunction<
  typeof createStudyGuideForCourse
>;

const COURSE_A = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";

function makeEnrollments(courseId = COURSE_A): EnrollmentResponse[] {
  return [
    {
      section: {
        id: "11111111-1111-1111-1111-111111111111",
        term: "Spring 2026",
        section_code: "01",
        instructor_name: "Dr. Aragoneses",
      },
      course: {
        id: courseId,
        department: "CPTS",
        number: "322",
        title: "Systems Programming",
      },
      school: { id: "s_1", acronym: "WSU" },
      role: "student",
      joined_at: "2026-04-20T10:00:00Z",
    },
  ];
}

beforeEach(() => {
  jest.clearAllMocks();
});

describe("NewStudyGuideForm (ASK-191)", () => {
  it("submits a valid draft to the selected course and redirects (AC1)", async () => {
    createMock.mockResolvedValue({
      id: "g_new",
    } as Awaited<ReturnType<typeof createStudyGuideForCourse>>);
    render(
      <NewStudyGuideForm
        enrollments={makeEnrollments()}
        defaultCourseId={null}
      />,
    );
    await userEvent.type(screen.getByLabelText(/title/i), "Buffer Overflows");
    await userEvent.type(
      screen.getByLabelText(/content/i),
      "Stack smashing walkthrough...",
    );
    await userEvent.click(
      screen.getByRole("button", { name: /save as draft/i }),
    );
    await waitFor(() => expect(createMock).toHaveBeenCalledTimes(1));
    expect(createMock).toHaveBeenCalledWith(
      COURSE_A,
      expect.objectContaining({
        title: "Buffer Overflows",
        content: "Stack smashing walkthrough...",
      }),
    );
    await waitFor(() =>
      expect(pushMock).toHaveBeenCalledWith("/study-guides/g_new"),
    );
  });

  it("toasts a forbidden ApiError without redirecting (AC2)", async () => {
    const apiErr = new ApiError(
      "POST /courses/.../study-guides failed: 403",
      { status: 403 } as unknown as Response,
      { code: 403, status: "forbidden", message: "Not enrolled" },
    );
    createMock.mockRejectedValue(apiErr);
    render(
      <NewStudyGuideForm
        enrollments={makeEnrollments()}
        defaultCourseId={null}
      />,
    );
    await userEvent.type(screen.getByLabelText(/title/i), "Buffer Overflows");
    await userEvent.type(
      screen.getByLabelText(/content/i),
      "Stack smashing walkthrough...",
    );
    await userEvent.click(
      screen.getByRole("button", { name: /save as draft/i }),
    );
    await waitFor(() => expect(toastErrorMock).toHaveBeenCalledWith(apiErr));
    expect(pushMock).not.toHaveBeenCalled();
    expect((screen.getByLabelText(/title/i) as HTMLInputElement).value).toBe(
      "Buffer Overflows",
    );
  });

  it("renders the empty state with a /courses link when no enrollments (AC3)", () => {
    render(<NewStudyGuideForm enrollments={[]} defaultCourseId={null} />);
    expect(screen.getByText("Join a course first")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Browse courses" }),
    ).toHaveAttribute("href", "/courses");
    expect(screen.queryByLabelText(/title/i)).not.toBeInTheDocument();
  });

  it("calls router.back() when Cancel is clicked (AC4)", async () => {
    render(
      <NewStudyGuideForm
        enrollments={makeEnrollments()}
        defaultCourseId={null}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
    expect(backMock).toHaveBeenCalledTimes(1);
  });
});
