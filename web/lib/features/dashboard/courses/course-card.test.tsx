/**
 * Exercises the ASK-180 acceptance criteria: the row variant surfaces
 * department + number + title ("CPTS 322 — Systems Programming" when read
 * top-to-bottom); the tile variant additionally shows the school name;
 * and the rightSlot is rendered without triggering the outer Link when
 * clicked.
 */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import type { CourseResponse } from "@/lib/api/types";

import { CourseCard } from "./course-card";

function makeCourse(overrides: Partial<CourseResponse> = {}): CourseResponse {
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
    ...overrides,
  };
}

describe("CourseCard / row variant", () => {
  it("shows department + number and title", () => {
    render(<CourseCard course={makeCourse()} variant="row" />);
    expect(screen.getByText("CPTS 322")).toBeInTheDocument();
    expect(screen.getByText("Systems Programming")).toBeInTheDocument();
  });

  it("links to /courses/{id} by default", () => {
    render(<CourseCard course={makeCourse()} variant="row" />);
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/courses/c_preview_1",
    );
  });

  it("honors a custom href", () => {
    render(
      <CourseCard
        course={makeCourse()}
        variant="row"
        href="/catalog?department=CPTS"
      />,
    );
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/catalog?department=CPTS",
    );
  });

  it("renders rightSlot and prevents clicks on it from navigating", async () => {
    const onSlotClick = jest.fn();
    render(
      <CourseCard
        course={makeCourse()}
        variant="row"
        rightSlot={
          <button type="button" onClick={onSlotClick}>
            Join
          </button>
        }
      />,
    );
    // The outer Link is still focusable/navigable, but clicks inside the
    // slot must not bubble to it -- otherwise an "Unenroll" affordance
    // would also open the course page, which would be confusing.
    await userEvent.click(screen.getByRole("button", { name: "Join" }));
    expect(onSlotClick).toHaveBeenCalledTimes(1);
  });
});

describe("CourseCard / tile variant", () => {
  it("renders department, title, and school name", () => {
    render(<CourseCard course={makeCourse()} variant="tile" />);
    expect(
      screen.getByRole("heading", { name: "CPTS 322" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Systems Programming")).toBeInTheDocument();
    expect(screen.getByText("Washington State University")).toBeInTheDocument();
  });

  it("links to /courses/{id} by default", () => {
    render(<CourseCard course={makeCourse()} variant="tile" />);
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/courses/c_preview_1",
    );
  });

  it("still renders rightSlot without navigating when clicked", async () => {
    const onSlotClick = jest.fn();
    render(
      <CourseCard
        course={makeCourse()}
        variant="tile"
        rightSlot={
          <button type="button" onClick={onSlotClick}>
            Joined
          </button>
        }
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Joined" }));
    expect(onSlotClick).toHaveBeenCalledTimes(1);
  });
});
