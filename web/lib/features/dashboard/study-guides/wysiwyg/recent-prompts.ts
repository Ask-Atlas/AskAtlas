/**
 * Persists the last few "Ask AI" prompts in localStorage so the user
 * can re-pick a directive they ran a minute ago without re-typing.
 *
 * SSR-safe: every helper checks for `window` before touching storage,
 * so the popover can render server-side without crashing.
 */

const STORAGE_KEY = "askatlas:study-guide-edit:recent-prompts";
const MAX_RECENT = 5;

function readStorage(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((v): v is string => typeof v === "string");
  } catch {
    // Quota errors, JSON parse failures, or a malicious extension that
    // overwrote the value -- prompt history is a best-effort UX nicety,
    // never let it break the editor.
    return [];
  }
}

function writeStorage(values: readonly string[]): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(values));
  } catch {
    // ignore quota / private-mode errors
  }
}

/** Returns the most-recent-first list of prompts, capped at {@link MAX_RECENT}. */
export function getRecentPrompts(): string[] {
  return readStorage().slice(0, MAX_RECENT);
}

/**
 * Inserts `prompt` at the head of the recent list, dedupes by exact
 * string match (case-insensitive trimmed compare), and trims to
 * {@link MAX_RECENT}. No-op for empty / whitespace-only strings.
 */
export function addRecentPrompt(prompt: string): string[] {
  const trimmed = prompt.trim();
  if (!trimmed) return getRecentPrompts();
  const existing = readStorage();
  const key = trimmed.toLowerCase();
  const deduped = existing.filter((v) => v.trim().toLowerCase() !== key);
  const next = [trimmed, ...deduped].slice(0, MAX_RECENT);
  writeStorage(next);
  return next;
}

export const RECENT_PROMPTS_STORAGE_KEY = STORAGE_KEY;
export const RECENT_PROMPTS_LIMIT = MAX_RECENT;
