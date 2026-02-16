"use client";

import { useMemo } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight } from "lucide-react";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  useSidebar,
} from "@/components/ui/sidebar";
import type {
  NavItem,
  NavRenderableItem,
  NavRenderableSubItem,
} from "@/lib/features/dashboard/sidebar/nav-main.types";
import { buildRenderableNavItems } from "@/lib/features/dashboard/sidebar/nav-main.utils";

type NavLinkItem = Extract<NavRenderableItem, { kind: "link" }>;
type NavCollapsibleItem = Extract<NavRenderableItem, { kind: "collapsible" }>;
type NavSectionItem = Extract<NavRenderableSubItem, { kind: "section" }>;
type NavSubLinkItem = Extract<NavRenderableSubItem, { kind: "link" }>;

function PrimaryNavLink({ item }: { item: NavLinkItem }) {
  return (
    <SidebarMenuItem key={item.key}>
      <SidebarMenuButton tooltip={item.title} asChild isActive={item.isActive}>
        <Link href={item.url}>
          {item.icon && <item.icon />}
          <span>{item.title}</span>
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}

function SubNavLink({ item }: { item: NavSubLinkItem }) {
  return (
    <SidebarMenuSubItem key={item.key}>
      <SidebarMenuSubButton asChild size={item.size} className={item.className}>
        <Link href={item.url}>
          <span>{item.title}</span>
        </Link>
      </SidebarMenuSubButton>
    </SidebarMenuSubItem>
  );
}

function SubNavSection({ item }: { item: NavSectionItem }) {
  return (
    <li key={item.key}>
      <Collapsible defaultOpen={item.defaultOpen} className="group/subsection">
        <CollapsibleTrigger asChild>
          <button
            type="button"
            className="text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex w-full items-center gap-2 rounded-md px-2 py-1 text-xs font-medium"
          >
            <span>{item.title}</span>
            <ChevronRight className="ml-auto size-3 transition-transform duration-200 group-data-[state=open]/subsection:rotate-90" />
          </button>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <ul className="mt-1 space-y-1">
            {item.items.map((subLink) => (
              <SubNavLink key={subLink.key} item={subLink} />
            ))}
          </ul>
        </CollapsibleContent>
      </Collapsible>
    </li>
  );
}

function renderSubNavItem(item: NavRenderableSubItem) {
  if (item.kind === "separator") {
    return <li key={item.key} className="bg-sidebar-border my-1 h-px" />;
  }

  if (item.kind === "section") {
    return <SubNavSection key={item.key} item={item} />;
  }

  return <SubNavLink key={item.key} item={item} />;
}

function CollapsibleNavItem({ item }: { item: NavCollapsibleItem }) {
  return (
    <Collapsible
      key={item.key}
      asChild
      defaultOpen={item.defaultOpen}
      className="group/collapsible"
    >
      <SidebarMenuItem>
        <CollapsibleTrigger asChild>
          <SidebarMenuButton tooltip={item.title} isActive={item.isActive}>
            {item.icon && <item.icon />}
            <span>{item.title}</span>
            <ChevronRight className="ml-auto size-4 text-sidebar-foreground/40 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
          </SidebarMenuButton>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <SidebarMenuSub>
            {item.subItems.map((subItem) => renderSubNavItem(subItem))}
          </SidebarMenuSub>
        </CollapsibleContent>
      </SidebarMenuItem>
    </Collapsible>
  );
}

function renderNavItem(item: NavRenderableItem) {
  if (item.kind === "link") {
    return <PrimaryNavLink key={item.key} item={item} />;
  }

  return <CollapsibleNavItem key={item.key} item={item} />;
}

export function NavMain({
  items,
  groupLabel = "Navigation",
}: {
  items: NavItem[];
  groupLabel?: string;
}) {
  const pathname = usePathname();
  const { state } = useSidebar();
  const isCollapsed = state === "collapsed";
  const preparedItems = useMemo(
    () => buildRenderableNavItems(items, pathname, isCollapsed),
    [items, isCollapsed, pathname],
  );

  return (
    <SidebarGroup>
      <SidebarGroupLabel className="sr-only">{groupLabel}</SidebarGroupLabel>
      <SidebarMenu>
        {preparedItems.map((item) => renderNavItem(item))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
