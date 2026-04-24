import type { StorybookConfig } from "@storybook/nextjs-vite";

const config: StorybookConfig = {
  stories: [
    "../lib/**/*.stories.@(ts|tsx|mdx)",
    "../components/**/*.stories.@(ts|tsx|mdx)",
  ],
  addons: [
    "@storybook/addon-a11y",
    "@storybook/addon-docs",
    "@storybook/addon-vitest",
  ],
  framework: "@storybook/nextjs-vite",
  staticDirs: ["../public"],
  // Deployed Storybook lives under a GitHub Pages subpath
  // (https://ask-atlas.github.io/AskAtlas/storybook/), so the build
  // needs its asset URLs rewritten. Local dev sticks with the default
  // root base, preserving `pnpm storybook` on localhost:6006.
  viteFinal: async (config) => {
    const raw = process.env.STORYBOOK_BASE_URL?.trim();
    // Normalize: empty -> root, always trailing slash so Vite's `base`
    // produces correct relative resolution for every chunk.
    config.base = !raw ? "/" : raw.endsWith("/") ? raw : `${raw}/`;
    return config;
  },
};

export default config;
