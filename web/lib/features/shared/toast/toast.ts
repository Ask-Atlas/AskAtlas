import { toast as sonnerToast } from "sonner";

import { ApiError } from "@/lib/api/errors";

export const toast = {
  success: (message: string): void => {
    sonnerToast.success(message);
  },
  error: (err: unknown): void => {
    if (err instanceof ApiError) {
      sonnerToast.error(err.body?.message ?? `Request failed (${err.status})`);
      return;
    }
    if (err instanceof Error) {
      sonnerToast.error(err.message);
      return;
    }
    sonnerToast.error("Something went wrong");
  },
  info: (message: string): void => {
    sonnerToast.info(message);
  },
};
