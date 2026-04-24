import type { RefSummary, RefsResolveResponse } from "@/lib/api/types";

const TOKEN = process.env.REFS_LIVE_TOKEN;
const BASE = process.env.REFS_LIVE_BASE ?? "https://api-staging.askatlas.study";

const describeLive = TOKEN ? describe : describe.skip;

async function apiGet<T>(path: string): Promise<T> {
  const resp = await fetch(`${BASE}${path}`, {
    headers: { Authorization: `Bearer ${TOKEN}` },
  });
  if (!resp.ok) throw new Error(`${path} returned ${resp.status}`);
  return (await resp.json()) as T;
}

async function resolve(
  refs: Array<{ type: string; id: string }>,
): Promise<RefsResolveResponse> {
  const resp = await fetch(`${BASE}/api/refs/resolve`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${TOKEN}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ refs }),
  });
  if (!resp.ok) throw new Error(`refs/resolve returned ${resp.status}`);
  return (await resp.json()) as RefsResolveResponse;
}

describeLive("refs-live: staging /api/refs/resolve round-trips", () => {
  jest.setTimeout(15_000);

  it("hydrates a real study guide with the expected SgRefSummary fields", async () => {
    const list = await apiGet<{ study_guides?: Array<{ id: string }> }>(
      "/api/me/study-guides?page_limit=1",
    );
    const sg = list.study_guides?.[0];
    if (!sg) {
      expect(true).toBe(true);
      return;
    }
    const body = await resolve([{ type: "sg", id: sg.id }]);
    const summary = body.results[`sg:${sg.id}`] as RefSummary | null;
    expect(summary).not.toBeNull();
    if (!summary) return;
    expect(summary.type).toBe("sg");
    expect(summary.id).toBe(sg.id);
    expect(typeof summary.title).toBe("string");
    expect(summary.course).toBeDefined();
    expect(typeof summary.course?.department).toBe("string");
    expect(typeof summary.quiz_count).toBe("number");
    expect(typeof summary.is_recommended).toBe("boolean");
  });

  it("hydrates a real course with the expected CourseRefSummary fields", async () => {
    const list = await apiGet<{ courses?: Array<{ id: string }> }>(
      "/api/courses?page_limit=1",
    );
    const course = list.courses?.[0];
    if (!course) {
      expect(true).toBe(true);
      return;
    }
    const body = await resolve([{ type: "course", id: course.id }]);
    const summary = body.results[`course:${course.id}`] as RefSummary | null;
    expect(summary).not.toBeNull();
    if (!summary) return;
    expect(summary.type).toBe("course");
    expect(typeof summary.department).toBe("string");
    expect(typeof summary.number).toBe("string");
    expect(typeof summary.school?.acronym).toBe("string");
  });

  it("mixed batch with a NIL file UUID returns null for that slot", async () => {
    const list = await apiGet<{ study_guides?: Array<{ id: string }> }>(
      "/api/me/study-guides?page_limit=1",
    );
    const sg = list.study_guides?.[0];
    const NIL = "00000000-0000-0000-0000-000000000000";
    const refs = [
      { type: "file", id: NIL },
      ...(sg ? [{ type: "sg", id: sg.id }] : []),
    ];
    const body = await resolve(refs);
    expect(body.results[`file:${NIL}`]).toBeNull();
    if (sg) expect(body.results[`sg:${sg.id}`]).not.toBeNull();
  });
});
