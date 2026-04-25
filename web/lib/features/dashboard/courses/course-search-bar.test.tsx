import { useState } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import type { ListCoursesQuery, SchoolResponse } from "@/lib/api/types";

import { CourseSearchBar } from "./course-search-bar";

const SCHOOLS: SchoolResponse[] = [
  {
    id: "11111111-1111-1111-1111-111111111111",
    name: "Atlas University",
    acronym: "AU",
    created_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "22222222-2222-2222-2222-222222222222",
    name: "Bay State College",
    acronym: "BSC",
    created_at: "2026-01-01T00:00:00Z",
  },
];

function Harness({
  initial = {},
  onChangeSpy,
  departments,
}: {
  initial?: ListCoursesQuery;
  onChangeSpy: jest.Mock;
  departments?: readonly string[];
}) {
  const [value, setValue] = useState<ListCoursesQuery>(initial);
  return (
    <CourseSearchBar
      value={value}
      onChange={(next) => {
        onChangeSpy(next);
        setValue(next);
      }}
      schools={SCHOOLS}
      departments={departments}
    />
  );
}

describe("CourseSearchBar", () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("debounces typed input and emits q after 250ms (AC1)", () => {
    const onChangeSpy = jest.fn();
    render(<Harness onChangeSpy={onChangeSpy} />);
    const input = screen.getByLabelText("Search courses") as HTMLInputElement;

    fireEvent.change(input, { target: { value: "algo" } });
    expect(onChangeSpy).not.toHaveBeenCalled();

    act(() => {
      jest.advanceTimersByTime(249);
    });
    expect(onChangeSpy).not.toHaveBeenCalled();

    act(() => {
      jest.advanceTimersByTime(1);
    });
    expect(onChangeSpy).toHaveBeenCalledTimes(1);
    expect(onChangeSpy).toHaveBeenCalledWith({ q: "algo" });
  });

  it("emits school_id immediately when a school is picked (AC2)", () => {
    const onChangeSpy = jest.fn();
    render(<Harness initial={{ q: "algo" }} onChangeSpy={onChangeSpy} />);
    const select = screen.getByLabelText(
      "Filter by school",
    ) as HTMLSelectElement;

    fireEvent.change(select, { target: { value: SCHOOLS[1].id } });

    expect(onChangeSpy).toHaveBeenCalledTimes(1);
    expect(onChangeSpy).toHaveBeenCalledWith({
      q: "algo",
      school_id: SCHOOLS[1].id,
    });
  });

  it("clears school_id when 'All schools' is chosen", () => {
    const onChangeSpy = jest.fn();
    render(
      <Harness
        initial={{ school_id: SCHOOLS[0].id }}
        onChangeSpy={onChangeSpy}
      />,
    );
    const select = screen.getByLabelText(
      "Filter by school",
    ) as HTMLSelectElement;
    expect(select.value).toBe(SCHOOLS[0].id);

    fireEvent.change(select, { target: { value: "__all__" } });
    expect(onChangeSpy).toHaveBeenCalledWith({ school_id: undefined });
  });

  it("emits q: undefined (not empty string) when the search is cleared (AC3)", () => {
    const onChangeSpy = jest.fn();
    render(<Harness initial={{ q: "algo" }} onChangeSpy={onChangeSpy} />);
    const input = screen.getByLabelText("Search courses") as HTMLInputElement;
    expect(input.value).toBe("algo");

    fireEvent.click(screen.getByRole("button", { name: "Clear search" }));
    expect(input.value).toBe("");

    act(() => {
      jest.advanceTimersByTime(250);
    });

    expect(onChangeSpy).toHaveBeenCalledTimes(1);
    const [arg] = onChangeSpy.mock.calls[0];
    expect(arg).toEqual({ q: undefined });
    expect(Object.prototype.hasOwnProperty.call(arg, "q")).toBe(true);
    expect(arg.q).toBeUndefined();
  });

  it("renders department filter only when departments are provided", () => {
    const onChangeSpy = jest.fn();
    const { rerender } = render(<Harness onChangeSpy={onChangeSpy} />);
    expect(screen.queryByLabelText("Filter by department")).toBeNull();

    rerender(
      <Harness onChangeSpy={onChangeSpy} departments={["CPTS", "MATH"]} />,
    );
    const dept = screen.getByLabelText(
      "Filter by department",
    ) as HTMLSelectElement;
    fireEvent.change(dept, { target: { value: "CPTS" } });
    expect(onChangeSpy).toHaveBeenCalledWith({ department: "CPTS" });
  });
});
