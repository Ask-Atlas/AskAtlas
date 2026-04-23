import { toast as sonnerToast } from "sonner";

import { ApiError } from "@/lib/api/errors";

type ToastId = string | number;

export const toast = {
  success: (message: string): ToastId => sonnerToast.success(message),
  info: (message: string): ToastId => sonnerToast.info(message),
  error: (err: unknown): ToastId => {
    if (typeof err === "string" && err.length > 0) {
      return sonnerToast.error(err);
    }
    if (err instanceof ApiError) {
      return sonnerToast.error(
        err.body?.message || `Request failed (${err.status})`,
      );
    }
    if (err instanceof Error && err.message) {
      return sonnerToast.error(err.message);
    }
    return sonnerToast.error("Something went wrong");
  },
  dismiss: (id?: ToastId): void => {
    sonnerToast.dismiss(id);
  },
};
