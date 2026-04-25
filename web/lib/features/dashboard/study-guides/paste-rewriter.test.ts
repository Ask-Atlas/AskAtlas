import { rewritePastedUrl } from "./paste-rewriter";

const HOSTS = ["askatlas.app", "www.askatlas.app", "staging.askatlas.study"];
const SG_ID = "11111111-2222-3333-4444-555555555555";
const QUIZ_ID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee";
const FILE_ID = "66666666-7777-8888-9999-000000000000";
const COURSE_ID = "cccccccc-dddd-eeee-ffff-111111111111";

describe("rewritePastedUrl", () => {
  it("rewrites a study-guide URL", () => {
    const r = rewritePastedUrl(
      `https://askatlas.app/study-guides/${SG_ID}`,
      HOSTS,
    );
    expect(r).toEqual({
      directive: `::sg{id="${SG_ID}"}`,
      type: "sg",
      id: SG_ID,
    });
  });

  it("rewrites a quiz URL", () => {
    const r = rewritePastedUrl(
      `https://askatlas.app/quizzes/${QUIZ_ID}`,
      HOSTS,
    );
    expect(r?.type).toBe("quiz");
    expect(r?.directive).toBe(`::quiz{id="${QUIZ_ID}"}`);
  });

  it("rewrites a file URL", () => {
    const r = rewritePastedUrl(`https://askatlas.app/files/${FILE_ID}`, HOSTS);
    expect(r?.type).toBe("file");
    expect(r?.directive).toBe(`::file{id="${FILE_ID}"}`);
  });

  it("rewrites a course URL", () => {
    const r = rewritePastedUrl(
      `https://askatlas.app/courses/${COURSE_ID}`,
      HOSTS,
    );
    expect(r?.type).toBe("course");
    expect(r?.directive).toBe(`::course{id="${COURSE_ID}"}`);
  });

  it("accepts a relative path", () => {
    const r = rewritePastedUrl(`/study-guides/${SG_ID}`, HOSTS);
    expect(r?.type).toBe("sg");
  });

  it("accepts any configured allowed host", () => {
    const r = rewritePastedUrl(
      `https://staging.askatlas.study/study-guides/${SG_ID}`,
      HOSTS,
    );
    expect(r?.type).toBe("sg");
  });

  it("lowercases the uuid in the emitted directive", () => {
    const upper = SG_ID.toUpperCase();
    const r = rewritePastedUrl(
      `https://askatlas.app/study-guides/${upper}`,
      HOSTS,
    );
    expect(r?.directive).toBe(`::sg{id="${SG_ID}"}`);
  });

  it("rejects external hosts", () => {
    expect(
      rewritePastedUrl(`https://evil.example/study-guides/${SG_ID}`, HOSTS),
    ).toBeNull();
  });

  it("rejects protocol-relative URLs pointing at an external host", () => {
    expect(
      rewritePastedUrl(`//evil.example/study-guides/${SG_ID}`, HOSTS),
    ).toBeNull();
  });

  it("rejects multi-URL paste (whitespace)", () => {
    const two = `https://askatlas.app/study-guides/${SG_ID} https://askatlas.app/quizzes/${QUIZ_ID}`;
    expect(rewritePastedUrl(two, HOSTS)).toBeNull();
  });

  it("rejects unknown entity routes", () => {
    expect(
      rewritePastedUrl(`https://askatlas.app/users/${SG_ID}`, HOSTS),
    ).toBeNull();
  });

  it("rejects a route with a non-uuid segment", () => {
    expect(
      rewritePastedUrl(`https://askatlas.app/study-guides/not-a-uuid`, HOSTS),
    ).toBeNull();
  });

  it("rejects deeper paths (e.g. /study-guides/<id>/edit)", () => {
    expect(
      rewritePastedUrl(
        `https://askatlas.app/study-guides/${SG_ID}/edit`,
        HOSTS,
      ),
    ).toBeNull();
  });

  it("rejects empty / non-URL strings", () => {
    expect(rewritePastedUrl("", HOSTS)).toBeNull();
    expect(rewritePastedUrl("  ", HOSTS)).toBeNull();
    expect(rewritePastedUrl("hello world", HOSTS)).toBeNull();
  });
});
