import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { fn } from "storybook/test";

import { ApiError } from "@/lib/api/errors";
import type { AppError } from "@/lib/api/types";

import { ErrorBoundaryContent } from "./error-boundary-content";

function fakeResponse(status: number): Response {
  return { status } as unknown as Response;
}

function apiError(status: number, body: AppError | null): ApiError {
  return new ApiError(`synthetic ${status}`, fakeResponse(status), body);
}

const meta: Meta<typeof ErrorBoundaryContent> = {
  title: "Shared/ErrorBoundaryContent",
  component: ErrorBoundaryContent,
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Client error boundary rendered by `app/error.tsx` and `app/(dashboard)/error.tsx`. Surfaces `ApiError.body.message`, falls back to a generic string, and hard-redirects to `/sign-in` on 401 (not demo-able in Storybook because the redirect would navigate the preview iframe away -- see unit tests for that branch).",
      },
    },
  },
};

export default meta;
type Story = StoryObj<typeof ErrorBoundaryContent>;

export const WithApiErrorBody: Story = {
  args: {
    error: apiError(403, {
      code: 403,
      status: "forbidden",
      message: "You don't have access to this resource.",
    }),
    reset: fn(),
  },
};

export const ApiErrorWithoutBody: Story = {
  args: {
    error: apiError(500, null),
    reset: fn(),
  },
};

export const GenericError: Story = {
  args: {
    error: new Error("Unexpected render failure"),
    reset: fn(),
  },
};
