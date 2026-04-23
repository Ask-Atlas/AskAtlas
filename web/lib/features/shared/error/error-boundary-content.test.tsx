/**
 * Covers the narrowing logic that matters in prod: 401 triggers a
 * hard redirect via the `hardRedirect` helper and renders nothing,
 * ApiError bodies get surfaced, everything else falls back to a safe
 * generic string. The retry button just has to wire to the `reset`
 * prop.
 */
jest.mock("./redirect", () => ({
  hardRedirect: jest.fn(),
}));

import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import { ApiError } from "@/lib/api/errors";
import type { AppError } from "@/lib/api/types";

import { ErrorBoundaryContent } from "./error-boundary-content";
import { hardRedirect } from "./redirect";

const mockRedirect = hardRedirect as jest.Mock;

function fakeResponse(status: number): Response {
  return { status } as unknown as Response;
}

describe("ErrorBoundaryContent", () => {
  beforeEach(() => {
    mockRedirect.mockClear();
  });

  it("redirects to /sign-in and renders nothing when the error is a 401", async () => {
    const { container } = render(
      <ErrorBoundaryContent
        error={new ApiError("op: 401", fakeResponse(401), null)}
        reset={jest.fn()}
      />,
    );

    await waitFor(() => expect(mockRedirect).toHaveBeenCalledWith("/sign-in"));
    expect(container).toBeEmptyDOMElement();
  });

  it("surfaces ApiError.body.message when the envelope is present", () => {
    const body: AppError = {
      code: 403,
      status: "forbidden",
      message: "You don't have access to this resource.",
    };
    render(
      <ErrorBoundaryContent
        error={new ApiError("op: 403", fakeResponse(403), body)}
        reset={jest.fn()}
      />,
    );

    expect(
      screen.getByText("You don't have access to this resource."),
    ).toBeInTheDocument();
    expect(mockRedirect).not.toHaveBeenCalled();
  });

  it("falls back to a generic message for ApiError without a body", () => {
    render(
      <ErrorBoundaryContent
        error={new ApiError("op: 500", fakeResponse(500), null)}
        reset={jest.fn()}
      />,
    );

    expect(
      screen.getByText("We hit an unexpected problem."),
    ).toBeInTheDocument();
  });

  it("falls back to a generic message for a plain Error", () => {
    render(
      <ErrorBoundaryContent error={new Error("boom")} reset={jest.fn()} />,
    );

    expect(
      screen.getByText("We hit an unexpected problem."),
    ).toBeInTheDocument();
  });

  it("invokes reset when the retry button is clicked", async () => {
    const reset = jest.fn();
    render(<ErrorBoundaryContent error={new Error("boom")} reset={reset} />);

    await userEvent.click(screen.getByRole("button", { name: /try again/i }));
    expect(reset).toHaveBeenCalledTimes(1);
  });
});
