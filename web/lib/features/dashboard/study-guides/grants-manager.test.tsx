/**
 * Exercises the ASK-212 optimistic-update contract for the
 * study-guide grants manager: add-success, add-failure (rollback +
 * toast), remove-success, remove-failure (rollback + toast). The
 * search debounce is respected by advancing timers in the tests that
 * care.
 */
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

jest.mock("../../shared/toast/toast", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
    info: jest.fn(),
    dismiss: jest.fn(),
  },
}));

import { toast } from "../../shared/toast/toast";
import type { StudyGuideGrantResponse } from "@/lib/api/types";

import { GrantsManager, type GrantsManagerActions } from "./grants-manager";

const mockList = jest.fn() as jest.MockedFunction<
  GrantsManagerActions["listGrants"]
>;
const mockCreate = jest.fn() as jest.MockedFunction<
  GrantsManagerActions["createGrant"]
>;
const mockRevoke = jest.fn() as jest.MockedFunction<
  GrantsManagerActions["revokeGrant"]
>;
const mockEnrollments = jest.fn() as jest.MockedFunction<
  GrantsManagerActions["listEnrollments"]
>;
const mockToastError = toast.error as jest.MockedFunction<typeof toast.error>;

const actions: GrantsManagerActions = {
  listGrants: (id) => mockList(id),
  listEnrollments: () => mockEnrollments(),
  createGrant: (id, body) => mockCreate(id, body),
  revokeGrant: (id, body) => mockRevoke(id, body),
};

const STUDY_GUIDE_ID = "sg_1";
const COURSE_ID_MATH = "course_math_340";
const COURSE_ID_CS = "course_cs_101";

function enrollmentFixture() {
  return {
    enrollments: [
      {
        section: {
          id: "sec_math_a",
          term: "Fall 2026",
          section_code: "A",
          instructor_name: "Prof. Euler",
        },
        course: {
          id: COURSE_ID_MATH,
          department: "MATH",
          number: "340",
          title: "Linear Algebra",
        },
        school: { id: "sch_1", acronym: "WSU" },
        role: "student" as const,
        joined_at: "2026-01-01T00:00:00Z",
      },
      {
        section: {
          id: "sec_cs_a",
          term: "Fall 2026",
          section_code: "A",
          instructor_name: "Prof. Lovelace",
        },
        course: {
          id: COURSE_ID_CS,
          department: "CS",
          number: "101",
          title: "Intro to CS",
        },
        school: { id: "sch_1", acronym: "WSU" },
        role: "student" as const,
        joined_at: "2026-01-01T00:00:00Z",
      },
    ],
  };
}

function grantFixture(
  overrides: Partial<StudyGuideGrantResponse> = {},
): StudyGuideGrantResponse {
  return {
    id: "grant_1",
    study_guide_id: STUDY_GUIDE_ID,
    grantee_type: "course",
    grantee_id: COURSE_ID_MATH,
    permission: "view",
    granted_by: "user_creator",
    created_at: "2026-04-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  jest.clearAllMocks();
  mockList.mockResolvedValue({ grants: [] });
  mockEnrollments.mockResolvedValue(enrollmentFixture());
});

describe("GrantsManager / initial load", () => {
  it("renders existing grants after the fetch resolves", async () => {
    mockList.mockResolvedValue({
      grants: [grantFixture({ grantee_id: COURSE_ID_MATH })],
    });
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    expect(
      await screen.findByTestId(`grant-chip-${COURSE_ID_MATH}`),
    ).toBeInTheDocument();
    expect(mockList).toHaveBeenCalledWith(STUDY_GUIDE_ID);
  });

  it("surfaces a toast when the initial fetch fails", async () => {
    mockList.mockRejectedValue(new Error("boom"));
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    await waitFor(() => expect(mockToastError).toHaveBeenCalled());
  });
});

describe("GrantsManager / add course grant", () => {
  it("inserts the chip optimistically before the network resolves", async () => {
    jest.useFakeTimers();
    let resolveCreate!: (value: StudyGuideGrantResponse) => void;
    mockCreate.mockImplementation(
      () =>
        new Promise<StudyGuideGrantResponse>((resolve) => {
          resolveCreate = resolve;
        }),
    );

    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    // Wait for the initial load to settle so the search input is live.
    await waitFor(() => expect(mockList).toHaveBeenCalled());

    await user.type(
      screen.getByRole("textbox", { name: /search courses or people/i }),
      "math",
    );
    act(() => {
      jest.advanceTimersByTime(200);
    });
    await user.click(await screen.findByText(/MATH 340/i));

    // Chip appears immediately -- the POST is still pending.
    expect(
      await screen.findByTestId(`grant-chip-${COURSE_ID_MATH}`),
    ).toBeInTheDocument();
    expect(mockCreate).toHaveBeenCalledWith(STUDY_GUIDE_ID, {
      grantee_type: "course",
      grantee_id: COURSE_ID_MATH,
      permission: "view",
    });

    // Let the pending promise resolve so act() stays happy.
    await act(async () => {
      resolveCreate?.({
        ...grantFixture({ grantee_id: COURSE_ID_MATH, id: "grant_new" }),
      });
    });
    jest.useRealTimers();
  });

  it("rolls back and fires a toast when the POST fails", async () => {
    jest.useFakeTimers();
    mockCreate.mockRejectedValue(new Error("nope"));
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    await waitFor(() => expect(mockList).toHaveBeenCalled());

    await user.type(
      screen.getByRole("textbox", { name: /search courses or people/i }),
      "cs",
    );
    act(() => {
      jest.advanceTimersByTime(200);
    });
    await user.click(await screen.findByText(/CS 101/i));

    // Chip first appears...
    expect(
      await screen.findByTestId(`grant-chip-${COURSE_ID_CS}`),
    ).toBeInTheDocument();
    // ...then disappears once the POST rejects.
    await waitFor(() =>
      expect(
        screen.queryByTestId(`grant-chip-${COURSE_ID_CS}`),
      ).not.toBeInTheDocument(),
    );
    expect(mockToastError).toHaveBeenCalled();
    jest.useRealTimers();
  });
});

describe("GrantsManager / remove grant", () => {
  it("removes the chip optimistically on click", async () => {
    mockList.mockResolvedValue({
      grants: [grantFixture({ grantee_id: COURSE_ID_MATH })],
    });
    let resolveRevoke: (() => void) | undefined;
    mockRevoke.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          resolveRevoke = resolve;
        }),
    );

    const user = userEvent.setup();
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    const chip = await screen.findByTestId(`grant-chip-${COURSE_ID_MATH}`);
    await user.click(within(chip).getByRole("button", { name: /remove/i }));
    expect(
      screen.queryByTestId(`grant-chip-${COURSE_ID_MATH}`),
    ).not.toBeInTheDocument();

    await act(async () => {
      resolveRevoke?.();
    });
  });

  it("restores the chip and fires a toast when the DELETE fails", async () => {
    mockList.mockResolvedValue({
      grants: [grantFixture({ grantee_id: COURSE_ID_MATH })],
    });
    mockRevoke.mockRejectedValue(new Error("nope"));

    const user = userEvent.setup();
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    const chip = await screen.findByTestId(`grant-chip-${COURSE_ID_MATH}`);
    await user.click(within(chip).getByRole("button", { name: /remove/i }));
    // Chip comes back once the DELETE rejects.
    expect(
      await screen.findByTestId(`grant-chip-${COURSE_ID_MATH}`),
    ).toBeInTheDocument();
    expect(mockToastError).toHaveBeenCalled();
  });
});

describe("GrantsManager / people search", () => {
  it("shows the coming-soon placeholder when the query starts with @", async () => {
    jest.useFakeTimers();
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    render(<GrantsManager studyGuideId={STUDY_GUIDE_ID} actions={actions} />);
    await waitFor(() => expect(mockList).toHaveBeenCalled());
    await user.type(
      screen.getByRole("textbox", { name: /search courses or people/i }),
      "@ada",
    );
    act(() => {
      jest.advanceTimersByTime(200);
    });
    expect(
      await screen.findByText(/people search coming soon/i),
    ).toBeInTheDocument();
    jest.useRealTimers();
  });
});
