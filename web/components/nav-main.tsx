"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight, type LucideIcon } from "lucide-react";

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

type NavSubLink = {
  title: string;
  url: string;
};

type NavSubSection = {
  type: "section";
  title: string;
  defaultOpen?: boolean;
  items: NavSubLink[];
};

type NavSeparator = {
  type: "separator";
};

type NavItem = {
  title: string;
  url: string;
  icon?: LucideIcon;
  items?: (NavSubLink | NavSubSection | NavSeparator)[];
};

function normalizePath(path: string) {
  if (path.length > 1 && path.endsWith("/")) {
    return path.slice(0, -1);
  }

  return path;
}

function matchesPath(pathname: string, target: string) {
  const normalizedPathname = normalizePath(pathname);
  const normalizedTarget = normalizePath(target);

  return (
    normalizedPathname === normalizedTarget ||
    normalizedPathname.startsWith(`${normalizedTarget}/`)
  );
}

function navItemIsActive(pathname: string, item: NavItem) {
  return matchesPath(pathname, item.url);
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

  return (
    <SidebarGroup>
      <SidebarGroupLabel className="sr-only">{groupLabel}</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => {
          const itemActive = navItemIsActive(pathname, item);

          return item.items?.length ? (
            isCollapsed ? (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  tooltip={item.title}
                  asChild
                  isActive={itemActive}
                >
                  <Link href={item.url}>
                    {item.icon && <item.icon />}
                    <span>{item.title}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ) : (
              <Collapsible
                key={item.title}
                asChild
                defaultOpen={itemActive}
                className="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton
                      tooltip={item.title}
                      isActive={itemActive}
                    >
                      {item.icon && <item.icon />}
                      <span>{item.title}</span>
                      <ChevronRight className="ml-auto size-4 text-sidebar-foreground/40 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {item.items.map((subItem, index) => {
                        if ("type" in subItem && subItem.type === "separator") {
                          return (
                            <li
                              key={`${item.title}-separator-${index}`}
                              className="bg-sidebar-border my-1 h-px"
                            />
                          );
                        }

                        if ("type" in subItem && subItem.type === "section") {
                          if (!subItem.items.length) {
                            return null;
                          }

                          const sectionActive = subItem.items.some(
                            (sectionItem) =>
                              matchesPath(pathname, sectionItem.url),
                          );

                          return (
                            <li key={`${item.title}-section-${subItem.title}`}>
                              <Collapsible
                                defaultOpen={
                                  sectionActive || subItem.defaultOpen
                                }
                                className="group/subsection"
                              >
                                <CollapsibleTrigger asChild>
                                  <button
                                    type="button"
                                    className="text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex w-full items-center gap-2 rounded-md px-2 py-1 text-xs font-medium"
                                  >
                                    <span>{subItem.title}</span>
                                    <ChevronRight className="ml-auto size-3 transition-transform duration-200 group-data-[state=open]/subsection:rotate-90" />
                                  </button>
                                </CollapsibleTrigger>
                                <CollapsibleContent>
                                  <ul className="mt-1 space-y-1">
                                    {subItem.items.map((sectionItem) => (
                                      <SidebarMenuSubItem
                                        key={sectionItem.title}
                                      >
                                        <SidebarMenuSubButton
                                          asChild
                                          size="sm"
                                          className="translate-x-0 pl-3 text-sidebar-foreground/65 hover:text-sidebar-foreground/85"
                                        >
                                          <Link href={sectionItem.url}>
                                            <span>{sectionItem.title}</span>
                                          </Link>
                                        </SidebarMenuSubButton>
                                      </SidebarMenuSubItem>
                                    ))}
                                  </ul>
                                </CollapsibleContent>
                              </Collapsible>
                            </li>
                          );
                        }

                        return (
                          <SidebarMenuSubItem key={subItem.title}>
                            <SidebarMenuSubButton asChild>
                              <Link href={subItem.url}>
                                <span>{subItem.title}</span>
                              </Link>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        );
                      })}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </SidebarMenuItem>
              </Collapsible>
            )
          ) : (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                tooltip={item.title}
                asChild
                isActive={itemActive}
              >
                <Link href={item.url}>
                  {item.icon && <item.icon />}
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          );
        })}
      </SidebarMenu>
    </SidebarGroup>
  );
}
