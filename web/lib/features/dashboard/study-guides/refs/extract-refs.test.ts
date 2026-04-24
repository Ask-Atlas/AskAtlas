import { extractRefs, refKey } from "./extract-refs";

describe("extractRefs", () => {
  const SG = "11111111-2222-3333-4444-555555555555";
  const QZ = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee";
  const FL = "66666666-7777-8888-9999-000000000000";
  const CR = "cccccccc-dddd-eeee-ffff-111111111111";

  it("finds inline sg refs", () => {
    const out = extractRefs(`Read ::sg{id="${SG}"} first.`);
    expect(out).toEqual([{ type: "sg", id: SG }]);
  });

  it("finds leaf refs in all four types", () => {
    const content = [
      `::sg{id="${SG}"}`,
      `::quiz{id="${QZ}"}`,
      `::file{id="${FL}"}`,
      `::course{id="${CR}"}`,
    ].join("\n");
    const out = extractRefs(content);
    expect(out).toHaveLength(4);
    expect(out).toEqual(
      expect.arrayContaining([
        { type: "sg", id: SG },
        { type: "quiz", id: QZ },
        { type: "file", id: FL },
        { type: "course", id: CR },
      ]),
    );
  });

  it("dedupes repeated refs", () => {
    const out = extractRefs(
      `::sg{id="${SG}"} ::sg{id="${SG}"} ::sg{id="${SG}"}`,
    );
    expect(out).toEqual([{ type: "sg", id: SG }]);
  });

  it("accepts single-quoted attribute values", () => {
    const out = extractRefs(`::quiz{id='${QZ}'}`);
    expect(out).toEqual([{ type: "quiz", id: QZ }]);
  });

  it("ignores directives with non-UUID ids", () => {
    expect(extractRefs(`::sg{id="not-a-uuid"}`)).toEqual([]);
  });

  it("ignores unknown directive names", () => {
    expect(extractRefs(`::user{id="${SG}"}`)).toEqual([]);
  });

  it("finds refs inside container directive bodies", () => {
    const content = `:::callout{type="note"}
See ::sg{id="${SG}"}.
:::`;
    expect(extractRefs(content)).toEqual([{ type: "sg", id: SG }]);
  });

  it("normalises uuid casing when deduping", () => {
    const lower = SG.toLowerCase();
    const upper = SG.toUpperCase();
    const out = extractRefs(`::sg{id="${lower}"} ::sg{id="${upper}"}`);
    expect(out).toHaveLength(1);
    expect(out[0].id).toBe(lower);
  });
});

describe("refKey", () => {
  it("formats as type:id lowercased", () => {
    expect(refKey("sg", "AAA")).toBe("sg:aaa");
  });
});
