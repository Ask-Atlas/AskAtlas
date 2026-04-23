import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/errors";

import { toast } from "./toast";
import { ToastProvider } from "./toast-provider";

function ToastPlayground() {
  return (
    <div className="flex flex-col items-start gap-3">
      <ToastProvider />
      <Button onClick={() => toast.success("Saved!")}>toast.success()</Button>
      <Button variant="secondary" onClick={() => toast.info("FYI")}>
        toast.info()
      </Button>
      <Button
        variant="destructive"
        onClick={() =>
          toast.error(
            new ApiError("DELETE /files/42: 403", { status: 403 } as Response, {
              code: 403,
              status: "forbidden",
              message: "You don't own this file.",
            }),
          )
        }
      >
        toast.error(ApiError 403)
      </Button>
      <Button variant="outline" onClick={() => toast.error("Network offline")}>
        toast.error(&quot;string&quot;)
      </Button>
      <Button variant="ghost" onClick={() => toast.error(new Error("boom"))}>
        toast.error(Error)
      </Button>
      <Button variant="ghost" onClick={() => toast.error(null)}>
        toast.error(unknown)
      </Button>
    </div>
  );
}

const meta: Meta<typeof ToastPlayground> = {
  title: "Shared/Toast",
  component: ToastPlayground,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Typed sonner wrapper. `toast.error(err)` narrows `ApiError` -> `err.body.message` (or `Request failed (STATUS)`), then `Error.message`, then the generic fallback. `toast.success`, `toast.info`, and `toast.dismiss` are passthrough.",
      },
    },
  },
};

export default meta;
type Story = StoryObj<typeof ToastPlayground>;

export const Playground: Story = {};
