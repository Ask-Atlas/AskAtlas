/**
 * Unit tests for the typed `toast` dispatcher. Exercises the error
 * narrowing so downstream callers can trust `toast.error(err)` with an
 * unknown value without leaking "[object Object]" style messages.
 *
 * `sonner` is mocked -- we only care that the right branch fires the
 * right message. Visual behaviour is covered by Playwright later.
 */
jest.mock("sonner", () => ({
  toast: {
    success: jest.fn(() => 1),
    error: jest.fn(() => 2),
    info: jest.fn(() => 3),
    dismiss: jest.fn(),
  },
}));

import { toast as sonnerToast } from "sonner";

import { ApiError } from "@/lib/api/errors";
import type { AppError } from "@/lib/api/types";

import { toast } from "./toast";

const mockSuccess = sonnerToast.success as jest.Mock;
const mockError = sonnerToast.error as jest.Mock;
const mockInfo = sonnerToast.info as jest.Mock;
const mockDismiss = sonnerToast.dismiss as jest.Mock;

function fakeResponse(status: number): Response {
  return { status } as unknown as Response;
}

beforeEach(() => {
  mockSuccess.mockClear();
  mockError.mockClear();
  mockInfo.mockClear();
  mockDismiss.mockClear();
});

describe("toast.success", () => {
  it("forwards the message and returns the sonner id", () => {
    const id = toast.success("Saved");
    expect(mockSuccess).toHaveBeenCalledWith("Saved");
    expect(id).toBe(1);
  });
});

describe("toast.info", () => {
  it("forwards the message and returns the sonner id", () => {
    const id = toast.info("Heads up");
    expect(mockInfo).toHaveBeenCalledWith("Heads up");
    expect(id).toBe(3);
  });
});

describe("toast.error", () => {
  it("surfaces a non-empty string as the message", () => {
    toast.error("bare string");
    expect(mockError).toHaveBeenCalledWith("bare string");
  });

  it("uses ApiError.body.message when the envelope is present", () => {
    const body: AppError = {
      code: 404,
      status: "not_found",
      message: "file missing",
    };
    toast.error(new ApiError("op failed: 404", fakeResponse(404), body));
    expect(mockError).toHaveBeenCalledWith("file missing");
  });

  it("falls back to `Request failed (STATUS)` when ApiError has no body", () => {
    toast.error(new ApiError("op failed: 500", fakeResponse(500), null));
    expect(mockError).toHaveBeenCalledWith("Request failed (500)");
  });

  it("falls back to `Request failed (STATUS)` when ApiError body.message is empty", () => {
    const body: AppError = { code: 500, status: "internal", message: "" };
    toast.error(new ApiError("op failed: 500", fakeResponse(500), body));
    expect(mockError).toHaveBeenCalledWith("Request failed (500)");
  });

  it("uses Error.message for non-ApiError errors", () => {
    toast.error(new Error("network down"));
    expect(mockError).toHaveBeenCalledWith("network down");
  });

  it("falls back to the generic message when Error.message is empty", () => {
    toast.error(new Error(""));
    expect(mockError).toHaveBeenCalledWith("Something went wrong");
  });

  it("falls back to the generic message for an empty string", () => {
    toast.error("");
    expect(mockError).toHaveBeenCalledWith("Something went wrong");
  });

  it("falls back to the generic message for null", () => {
    toast.error(null);
    expect(mockError).toHaveBeenCalledWith("Something went wrong");
  });

  it("returns the sonner id", () => {
    expect(toast.error(new Error("x"))).toBe(2);
  });
});

describe("toast.dismiss", () => {
  it("forwards a specific id to sonner.dismiss", () => {
    toast.dismiss(7);
    expect(mockDismiss).toHaveBeenCalledWith(7);
  });

  it("dismisses all toasts when called without an id", () => {
    toast.dismiss();
    expect(mockDismiss).toHaveBeenCalledWith(undefined);
  });
});
