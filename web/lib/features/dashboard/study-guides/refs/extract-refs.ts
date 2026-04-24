export const ENTITY_DIRECTIVE_NAMES = ["sg", "quiz", "file", "course"] as const;
export type EntityType = (typeof ENTITY_DIRECTIVE_NAMES)[number];

export interface EntityRef {
  type: EntityType;
  id: string;
}

const DIRECTIVE_RE =
  /:{1,3}(sg|quiz|file|course)\{[^}]*\bid=(?:"([^"]+)"|'([^']+)'|([^\s}]+))[^}]*\}/gi;
const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function extractRefs(content: string): EntityRef[] {
  const out = new Map<string, EntityRef>();
  for (const match of content.matchAll(DIRECTIVE_RE)) {
    const type = match[1].toLowerCase() as EntityType;
    const id = (match[2] ?? match[3] ?? match[4] ?? "").toLowerCase();
    if (!UUID_RE.test(id)) continue;
    const key = `${type}:${id}`;
    if (!out.has(key)) out.set(key, { type, id });
  }
  return Array.from(out.values());
}

export function refKey(type: EntityType, id: string): string {
  return `${type}:${id.toLowerCase()}`;
}
