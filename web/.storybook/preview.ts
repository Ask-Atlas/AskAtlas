import type { Preview } from "@storybook/nextjs-vite";

import "../app/globals.css";

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    a11y: {
      test: "todo",
    },
    layout: "centered",
    backgrounds: {
      default: "app",
      values: [
        { name: "app", value: "oklch(1 0 0)" },
        { name: "muted", value: "oklch(0.97 0.002 247)" },
      ],
    },
  },
};

export default preview;
