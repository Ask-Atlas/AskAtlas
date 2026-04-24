const UUID = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}";

const LEAF_RE = new RegExp(
  `::(sg|quiz|file|course)\\{\\s*id="(${UUID})"\\s*\\}`,
  "gi",
);
const INLINE_RE = new RegExp(
  `(?<!:):(sg|quiz|file|course)\\{\\s*id="(${UUID})"\\s*\\}`,
  "gi",
);

export function preprocessMarkdown(md: string): string {
  return md
    .replace(
      LEAF_RE,
      (_, type, id) => `<${type}-ref id="${id.toLowerCase()}"></${type}-ref>`,
    )
    .replace(
      INLINE_RE,
      (_, type, id) =>
        `<${type}-ref-inline id="${id.toLowerCase()}"></${type}-ref-inline>`,
    );
}
