const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

const ROUTES: Record<string, "sg" | "quiz" | "file" | "course"> = {
  "study-guides": "sg",
  quizzes: "quiz",
  files: "file",
  courses: "course",
};

export interface PasteRewriteMatch {
  directive: string;
  type: "sg" | "quiz" | "file" | "course";
  id: string;
}

function parse(url: string): URL | null {
  try {
    if (url.startsWith("/")) {
      return new URL(url, "http://placeholder.invalid");
    }
    return new URL(url);
  } catch {
    return null;
  }
}

function hostMatches(parsed: URL, allowedHosts: string[]): boolean {
  if (parsed.hostname === "placeholder.invalid") return true;
  return allowedHosts.includes(parsed.hostname);
}

export function rewritePastedUrl(
  pasted: string,
  allowedHosts: string[],
): PasteRewriteMatch | null {
  const trimmed = pasted.trim();
  if (trimmed === "") return null;
  if (/\s/.test(trimmed)) return null;

  const parsed = parse(trimmed);
  if (!parsed) return null;
  if (!hostMatches(parsed, allowedHosts)) return null;

  const segments = parsed.pathname.split("/").filter(Boolean);
  if (segments.length !== 2) return null;

  const [route, id] = segments;
  const type = ROUTES[route];
  if (!type) return null;
  if (!UUID_RE.test(id)) return null;

  return {
    directive: `::${type}{id="${id.toLowerCase()}"}`,
    type,
    id: id.toLowerCase(),
  };
}
