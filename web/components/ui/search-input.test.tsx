import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

import { SearchInput } from "./search-input";

function ControlledHarness(props: {
  initial?: string;
  withClear?: boolean;
  clearLabel?: string;
}) {
  const [value, setValue] = useState(props.initial ?? "");
  return (
    <SearchInput
      placeholder="Search…"
      value={value}
      onChange={(event) => setValue(event.target.value)}
      onClear={props.withClear ? () => setValue("") : undefined}
      clearLabel={props.clearLabel}
    />
  );
}

describe("SearchInput", () => {
  it("renders a controlled search input that reflects typed values", () => {
    render(<ControlledHarness />);
    const input = screen.getByPlaceholderText("Search…") as HTMLInputElement;
    expect(input.type).toBe("text");
    expect(input.getAttribute("role")).toBe("searchbox");
    fireEvent.change(input, { target: { value: "algo" } });
    expect(input.value).toBe("algo");
  });

  it("hides the clear button when value is empty even if onClear is provided", () => {
    render(<ControlledHarness withClear />);
    expect(
      screen.queryByRole("button", { name: "Clear search" }),
    ).not.toBeInTheDocument();
  });

  it("hides the clear button when onClear is omitted (uncleardable mode)", () => {
    render(<ControlledHarness initial="algo" />);
    expect(
      screen.queryByRole("button", { name: "Clear search" }),
    ).not.toBeInTheDocument();
  });

  it("shows the clear button when value is non-empty and clears on click", () => {
    render(<ControlledHarness initial="algo" withClear />);
    const input = screen.getByPlaceholderText("Search…") as HTMLInputElement;
    const clear = screen.getByRole("button", { name: "Clear search" });
    fireEvent.click(clear);
    expect(input.value).toBe("");
    expect(
      screen.queryByRole("button", { name: "Clear search" }),
    ).not.toBeInTheDocument();
  });

  it("honors a custom clearLabel for localization", () => {
    render(<ControlledHarness initial="algo" withClear clearLabel="Effacer" />);
    expect(screen.getByRole("button", { name: "Effacer" })).toBeInTheDocument();
  });
});
