import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import Home from "./page";

describe("Home", () => {
  it("renders the page", () => {
    render(<Home />);
    expect(screen.getByText(/To get started/i)).toBeInTheDocument();
  });
});
