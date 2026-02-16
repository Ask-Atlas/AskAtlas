import {
  buildRenderableNavItems,
  matchesPath,
  normalizePath,
  sanitizeSubItems,
} from "./nav-main.utils";
import type { NavItem, NavSubItem } from "./nav-main.types";

describe("nav-main.utils", () => {
  describe("normalizePath", () => {
    it("removes trailing slash except root", () => {
      expect(normalizePath("/courses/")).toBe("/courses");
      expect(normalizePath("/")).toBe("/");
    });
  });

  describe("matchesPath", () => {
    it("matches exact and nested paths", () => {
      expect(matchesPath("/courses", "/courses")).toBe(true);
      expect(matchesPath("/courses/intro", "/courses")).toBe(true);
      expect(matchesPath("/resources", "/courses")).toBe(false);
    });
  });

  describe("sanitizeSubItems", () => {
    it("removes leading, trailing, and duplicate separators", () => {
      const items: NavSubItem[] = [
        { type: "separator" },
        { title: "A", url: "/a" },
        { type: "separator" },
        { type: "separator" },
        { title: "B", url: "/b" },
        { type: "separator" },
      ];

      expect(sanitizeSubItems(items)).toEqual([
        { title: "A", url: "/a" },
        { type: "separator" },
        { title: "B", url: "/b" },
      ]);
    });

    it("removes empty sections", () => {
      const items: NavSubItem[] = [
        {
          type: "section",
          title: "Recent",
          items: [],
        },
      ];

      expect(sanitizeSubItems(items)).toEqual([]);
    });
  });

  describe("buildRenderableNavItems", () => {
    const navItems: NavItem[] = [
      {
        title: "Courses",
        url: "/courses",
        items: [{ title: "Browse", url: "/courses" }],
      },
      {
        title: "Resources",
        url: "/resources",
      },
    ];

    it("builds collapsible items when sidebar is expanded", () => {
      const rendered = buildRenderableNavItems(navItems, "/courses", false);

      expect(rendered[0]).toMatchObject({
        kind: "collapsible",
        isActive: true,
      });
      expect(rendered[1]).toMatchObject({
        kind: "link",
        isActive: false,
      });
    });

    it("builds plain links when sidebar is collapsed", () => {
      const rendered = buildRenderableNavItems(navItems, "/courses", true);

      expect(rendered[0]).toMatchObject({
        kind: "link",
        isActive: true,
      });
    });
  });
});
