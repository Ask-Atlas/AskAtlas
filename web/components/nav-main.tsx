"use client";

import Link from "next/link";
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
  isActive?: boolean;
  items?: (NavSubLink | NavSubSection | NavSeparator)[];
};

export function NavMain({ items }: { items: NavItem[] }) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel className="sr-only">Navigation</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) =>
          item.items?.length ? (
            <Collapsible
              key={item.title}
              asChild
              defaultOpen={item.isActive}
              className="group/collapsible"
            >
              <SidebarMenuItem>
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton tooltip={item.title}>
                    {item.icon && <item.icon />}
                    <span>{item.title}</span>
                    <ChevronRight className="ml-auto size-4 text-sidebar-foreground/40 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {item.items?.map((subItem, index) => {
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

                        return (
                          <li key={`${item.title}-section-${subItem.title}`}>
                            <Collapsible
                              defaultOpen={subItem.defaultOpen}
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
                                    <SidebarMenuSubItem key={sectionItem.title}>
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
          ) : (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                tooltip={item.title}
                asChild
                isActive={item.isActive}
              >
                <Link href={item.url}>
                  {item.icon && <item.icon />}
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ),
        )}
      </SidebarMenu>
    </SidebarGroup>
  );
}
