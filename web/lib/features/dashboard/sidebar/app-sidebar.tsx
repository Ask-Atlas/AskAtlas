"use client";

import * as React from "react";
import Link from "next/link";
import { BookOpen, FileText, Globe, Home, Library, Star } from "lucide-react";
import { useDashboardCommonCopy } from "@/lib/features/dashboard/i18n/common/common-copy-provider";

import { NavMain } from "@/lib/features/dashboard/sidebar/nav-main";
import { NavUser } from "@/lib/features/dashboard/sidebar/nav-user";
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

function getNavMain(
  copy: ReturnType<typeof useDashboardCommonCopy>,
): React.ComponentProps<typeof NavMain>["items"] {
  return [
    {
      title: copy.sidebar.items.home,
      url: "/home",
      icon: Home,
    },
    {
      title: copy.sidebar.items.starred,
      url: "/me/saved",
      icon: Star,
      items: [
        {
          title: copy.sidebar.items.starredAll,
          url: "/me/saved",
        },
        {
          type: "separator",
        },
        {
          type: "section",
          title: copy.sidebar.items.starredRecent,
          defaultOpen: true,
          items: [
            {
              title: copy.sidebar.items.samples.machineLearningFundamentals,
              url: "/courses/machine-learning-fundamentals",
            },
            {
              title: copy.sidebar.items.samples.binaryTreesCheatSheet,
              url: "/study-guides/binary-trees-cheat-sheet",
            },
            {
              title: copy.sidebar.items.samples.neuralNetworksPaper,
              url: "/resources/neural-networks-paper",
            },
          ],
        },
      ],
    },
    {
      title: copy.sidebar.items.courses,
      url: "/courses",
      icon: BookOpen,
      items: [
        {
          title: copy.sidebar.items.coursesBrowse,
          url: "/courses",
        },
        {
          title: copy.sidebar.items.coursesMine,
          url: "/me/courses",
        },
        {
          type: "separator",
        },
        {
          type: "section",
          title: copy.sidebar.items.coursesSectionMine,
          defaultOpen: true,
          items: [
            {
              title: copy.sidebar.items.samples.introPsychology,
              url: "/courses/introduction-to-psychology",
            },
            {
              title: copy.sidebar.items.samples.dataStructuresAlgorithms,
              url: "/courses/data-structures-and-algorithms",
            },
            {
              title: copy.sidebar.items.samples.modernWebDevelopment,
              url: "/courses/modern-web-development",
            },
            {
              title: copy.sidebar.items.samples.machineLearningFundamentals,
              url: "/courses/machine-learning-fundamentals",
            },
          ],
        },
      ],
    },
    {
      title: copy.sidebar.items.studyGuides,
      url: "/study-guides",
      icon: FileText,
      items: [
        {
          title: copy.sidebar.items.guidesCreate,
          url: "/study-guides/new",
        },
        {
          title: copy.sidebar.items.guidesMine,
          url: "/me/study-guides",
        },
        {
          type: "separator",
        },
        {
          type: "section",
          title: copy.sidebar.items.guidesRecent,
          defaultOpen: true,
          items: [
            {
              title: copy.sidebar.items.samples.midtermReviewPsychology,
              url: "/study-guides/midterm-review-psychology",
            },
            {
              title: copy.sidebar.items.samples.binaryTreesCheatSheet,
              url: "/study-guides/binary-trees-cheat-sheet",
            },
            {
              title: copy.sidebar.items.samples.algorithmComplexityNotes,
              url: "/study-guides/algorithm-complexity-notes",
            },
          ],
        },
      ],
    },
    {
      title: copy.sidebar.items.resources,
      url: "/resources",
      icon: Library,
      items: [
        {
          title: copy.sidebar.items.resourcesUpload,
          url: "/resources/upload",
        },
        {
          title: copy.sidebar.items.resourcesView,
          url: "/resources",
        },
        {
          type: "separator",
        },
        {
          type: "section",
          title: copy.sidebar.items.resourcesRecent,
          defaultOpen: true,
          items: [
            {
              title: copy.sidebar.items.samples.neuralNetworksPaper,
              url: "/resources/neural-networks-paper",
            },
            {
              title: copy.sidebar.items.samples.databasesQuickReference,
              url: "/resources/databases-quick-reference",
            },
            {
              title: copy.sidebar.items.samples.cloudComputingNotes,
              url: "/resources/cloud-computing-notes",
            },
          ],
        },
      ],
    },
  ];
}

const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatars/shadcn.jpg",
  },
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const copy = useDashboardCommonCopy();
  const navMain = getNavMain(copy);

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link href="/home">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-linear-to-br from-primary to-primary/70 text-primary-foreground">
                  <Globe className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">
                    {copy.sidebar.brandName}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {copy.sidebar.brandTagline}
                  </span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMain} groupLabel={copy.sidebar.groupLabel} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} labels={copy.userMenu} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
