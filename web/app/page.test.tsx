import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import Home from "./(marketing)/page";

describe("Home", () => {
  it("renders the page", () => {
    render(<Home />);
    expect(screen.getByText(/One app for every class/i)).toBeInTheDocument();
  });
});
