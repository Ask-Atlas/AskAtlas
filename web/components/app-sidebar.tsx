"use client";

import * as React from "react";
import Link from "next/link";
import { BookOpen, FileText, Globe, Home, Library, Star } from "lucide-react";

import { NavMain } from "@/components/nav-main";
import { NavUser } from "@/components/nav-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";

const navMain: React.ComponentProps<typeof NavMain>["items"] = [
  {
    title: "Home",
    url: "/home",
    icon: Home,
    isActive: true,
  },
  {
    title: "Starred",
    url: "/me/saved",
    icon: Star,
    items: [
      {
        title: "All Starred",
        url: "/me/saved",
      },
      {
        type: "separator",
      },
      {
        type: "section",
        title: "Recently Starred",
        defaultOpen: true,
        items: [
          {
            title: "Machine Learning Fundamentals",
            url: "/courses/machine-learning-fundamentals",
          },
          {
            title: "Binary Trees Cheat Sheet",
            url: "/study-guides/binary-trees-cheat-sheet",
          },
          {
            title: "Neural Networks Paper",
            url: "/resources/neural-networks-paper",
          },
        ],
      },
    ],
  },
  {
    title: "Courses",
    url: "/courses",
    icon: BookOpen,
    items: [
      {
        title: "Browse Courses",
        url: "/courses",
      },
      {
        title: "My Courses",
        url: "/me/courses",
      },
      {
        type: "separator",
      },
      {
        type: "section",
        title: "My Courses",
        defaultOpen: true,
        items: [
          {
            title: "Introduction to Psychology",
            url: "/courses/introduction-to-psychology",
          },
          {
            title: "Data Structures & Algorithms",
            url: "/courses/data-structures-and-algorithms",
          },
          {
            title: "Modern Web Development",
            url: "/courses/modern-web-development",
          },
          {
            title: "Machine Learning Fundamentals",
            url: "/courses/machine-learning-fundamentals",
          },
        ],
      },
    ],
  },
  {
    title: "Study Guides",
    url: "/study-guides",
    icon: FileText,
    items: [
      {
        title: "Create New Guide",
        url: "/study-guides/new",
      },
      {
        title: "My Study Guides",
        url: "/me/study-guides",
      },
      {
        type: "separator",
      },
      {
        type: "section",
        title: "Recently Viewed",
        defaultOpen: true,
        items: [
          {
            title: "Midterm Review - Psychology",
            url: "/study-guides/midterm-review-psychology",
          },
          {
            title: "Binary Trees Cheat Sheet",
            url: "/study-guides/binary-trees-cheat-sheet",
          },
          {
            title: "Algorithm Complexity Notes",
            url: "/study-guides/algorithm-complexity-notes",
          },
        ],
      },
    ],
  },
  {
    title: "Resources",
    url: "/resources",
    icon: Library,
    items: [
      {
        title: "Upload Resource",
        url: "/resources/upload",
      },
      {
        title: "View Resources",
        url: "/resources",
      },
      {
        type: "separator",
      },
      {
        type: "section",
        title: "Recently Viewed",
        defaultOpen: true,
        items: [
          {
            title: "Neural Networks Paper",
            url: "/resources/neural-networks-paper",
          },
          {
            title: "Databases Quick Reference",
            url: "/resources/databases-quick-reference",
          },
          {
            title: "Cloud Computing Notes",
            url: "/resources/cloud-computing-notes",
          },
        ],
      },
    ],
  },
];

// This is sample data.
const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatars/shadcn.jpg",
  },
  navMain,
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link href="/home">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-primary/70 text-primary-foreground">
                  <Globe className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">AskAtlas</span>
                  <span className="truncate text-xs text-muted-foreground">
                    Your study workspace
                  </span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
