import type {
  NavItem,
  NavRenderableItem,
  NavRenderableSubItem,
  NavSeparator,
  NavSubItem,
  NavSubSection,
  PreparedNavItem,
} from "./nav-main.types";

const MAX_SUBSECTION_ITEMS = 5;

export function normalizePath(path: string): string {
  if (path.length > 1 && path.endsWith("/")) {
    return path.slice(0, -1);
  }

  return path;
}

export function matchesPath(pathname: string, target: string): boolean {
  const normalizedPathname = normalizePath(pathname);
  const normalizedTarget = normalizePath(target);

  return (
    normalizedPathname === normalizedTarget ||
    normalizedPathname.startsWith(`${normalizedTarget}/`)
  );
}

export function isSeparator(item: NavSubItem): item is NavSeparator {
  return "type" in item && item.type === "separator";
}

export function isSection(item: NavSubItem): item is NavSubSection {
  return "type" in item && item.type === "section";
}

export function sanitizeSubItems(items?: NavSubItem[]): NavSubItem[] {
  if (!items?.length) {
    return [];
  }

  const normalizedItems = items
    .map((item) => {
      if (isSection(item)) {
        const limitedItems = item.items.slice(0, MAX_SUBSECTION_ITEMS);

        if (!limitedItems.length) {
          return null;
        }

        return {
          ...item,
          items: limitedItems,
        };
      }

      return item;
    })
    .filter((item): item is NavSubItem => item !== null);

  return normalizedItems.filter((item, index, list) => {
    if (!isSeparator(item)) {
      return true;
    }

    const isFirst = index === 0;
    const isLast = index === list.length - 1;
    const previousItem = index > 0 ? list[index - 1] : null;
    const previousIsSeparator = previousItem
      ? isSeparator(previousItem)
      : false;

    return !isFirst && !isLast && !previousIsSeparator;
  });
}

export function prepareNavItems(
  items: NavItem[],
  pathname: string,
): PreparedNavItem[] {
  return items.map((item) => ({
    ...item,
    isActive: matchesPath(pathname, item.url),
    subItems: sanitizeSubItems(item.items),
  }));
}

function buildRenderableSubItems(
  parentTitle: string,
  subItems: NavSubItem[],
  pathname: string,
): NavRenderableSubItem[] {
  return subItems.map((subItem, index) => {
    if (isSeparator(subItem)) {
      return {
        kind: "separator",
        key: `${parentTitle}-separator-${index}`,
      };
    }

    if (isSection(subItem)) {
      const sectionDefaultOpen =
        subItem.items.some((sectionItem) =>
          matchesPath(pathname, sectionItem.url),
        ) || Boolean(subItem.defaultOpen);

      return {
        kind: "section",
        key: `${parentTitle}-section-${subItem.title}`,
        title: subItem.title,
        defaultOpen: sectionDefaultOpen,
        items: subItem.items.map((sectionItem) => ({
          kind: "link",
          key: `${parentTitle}-section-item-${sectionItem.title}`,
          title: sectionItem.title,
          url: sectionItem.url,
          size: "sm",
          className:
            "translate-x-0 pl-3 text-sidebar-foreground/65 hover:text-sidebar-foreground/85",
        })),
      };
    }

    return {
      kind: "link",
      key: `${parentTitle}-item-${subItem.title}`,
      title: subItem.title,
      url: subItem.url,
      size: "md",
    };
  });
}

export function buildRenderableNavItems(
  items: NavItem[],
  pathname: string,
  isCollapsed: boolean,
): NavRenderableItem[] {
  return prepareNavItems(items, pathname).map((item) => {
    const hasSubItems = item.subItems.length > 0;

    if (!hasSubItems || isCollapsed) {
      return {
        kind: "link",
        key: item.title,
        title: item.title,
        url: item.url,
        icon: item.icon,
        isActive: item.isActive,
      };
    }

    return {
      kind: "collapsible",
      key: item.title,
      title: item.title,
      icon: item.icon,
      isActive: item.isActive,
      defaultOpen: item.isActive,
      subItems: buildRenderableSubItems(item.title, item.subItems, pathname),
    };
  });
}
