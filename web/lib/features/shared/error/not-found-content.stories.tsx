import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import Link from "next/link";

import { Button } from "@/components/ui/button";

import { NotFoundContent } from "./not-found-content";

const meta: Meta<typeof NotFoundContent> = {
  title: "Shared/NotFoundContent",
  component: NotFoundContent,
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "404 body used by both the root and `(dashboard)` `not-found.tsx` routes. Takes an `action` slot so each wrapper can pick a context-appropriate CTA.",
      },
    },
  },
};

export default meta;
type Story = StoryObj<typeof NotFoundContent>;

export const Authenticated: Story = {
  args: {
    action: (
      <Button asChild>
        <Link href="/home">Back to dashboard</Link>
      </Button>
    ),
  },
};

export const Unauthenticated: Story = {
  args: {
    action: (
      <Button asChild>
        <Link href="/">Back to home</Link>
      </Button>
    ),
  },
};
