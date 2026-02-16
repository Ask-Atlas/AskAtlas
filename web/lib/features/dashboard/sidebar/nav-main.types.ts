import type { LucideIcon } from "lucide-react";

export type NavSubLink = {
  title: string;
  url: string;
};

export type NavSubSection = {
  type: "section";
  title: string;
  defaultOpen?: boolean;
  items: NavSubLink[];
};

export type NavSeparator = {
  type: "separator";
};

export type NavSubItem = NavSubLink | NavSubSection | NavSeparator;

export type NavItem = {
  title: string;
  url: string;
  icon?: LucideIcon;
  items?: NavSubItem[];
};

export type PreparedNavItem = NavItem & {
  isActive: boolean;
  subItems: NavSubItem[];
};

export type NavRenderableSubLink = {
  kind: "link";
  key: string;
  title: string;
  url: string;
  size?: "sm" | "md";
  className?: string;
};

export type NavRenderableSubSection = {
  kind: "section";
  key: string;
  title: string;
  defaultOpen: boolean;
  items: NavRenderableSubLink[];
};

export type NavRenderableSeparator = {
  kind: "separator";
  key: string;
};

export type NavRenderableSubItem =
  | NavRenderableSubLink
  | NavRenderableSubSection
  | NavRenderableSeparator;

export type NavRenderableItemLink = {
  kind: "link";
  key: string;
  title: string;
  url: string;
  icon?: LucideIcon;
  isActive: boolean;
};

export type NavRenderableItemCollapsible = {
  kind: "collapsible";
  key: string;
  title: string;
  icon?: LucideIcon;
  isActive: boolean;
  defaultOpen: boolean;
  subItems: NavRenderableSubItem[];
};

export type NavRenderableItem =
  | NavRenderableItemLink
  | NavRenderableItemCollapsible;
