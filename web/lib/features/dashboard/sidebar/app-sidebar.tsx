"use client";

import * as React from "react";
import Link from "next/link";
import { BookOpen, FileText, Globe, Home, Library, Star } from "lucide-react";
import { useDashboardCommonCopy } from "@/lib/features/dashboard/i18n/common/common-copy-provider";
import { DASHBOARD_ROUTES } from "@/lib/features/dashboard/navigation/routes";

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
      url: DASHBOARD_ROUTES.home,
      icon: Home,
    },
    {
      title: copy.sidebar.items.starred,
      url: DASHBOARD_ROUTES.saved,
      icon: Star,
      items: [
        {
          title: copy.sidebar.items.starredAll,
          url: DASHBOARD_ROUTES.saved,
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
              url: DASHBOARD_ROUTES.samples.machineLearningFundamentals,
            },
            {
              title: copy.sidebar.items.samples.binaryTreesCheatSheet,
              url: DASHBOARD_ROUTES.samples.binaryTreesCheatSheet,
            },
            {
              title: copy.sidebar.items.samples.neuralNetworksPaper,
              url: DASHBOARD_ROUTES.samples.neuralNetworksPaper,
            },
          ],
        },
      ],
    },
    {
      title: copy.sidebar.items.courses,
      url: DASHBOARD_ROUTES.courses,
      icon: BookOpen,
      items: [
        {
          title: copy.sidebar.items.coursesBrowse,
          url: DASHBOARD_ROUTES.courses,
        },
        {
          title: copy.sidebar.items.coursesMine,
          url: DASHBOARD_ROUTES.myCourses,
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
              url: DASHBOARD_ROUTES.samples.introPsychology,
            },
            {
              title: copy.sidebar.items.samples.dataStructuresAlgorithms,
              url: DASHBOARD_ROUTES.samples.dataStructuresAlgorithms,
            },
            {
              title: copy.sidebar.items.samples.modernWebDevelopment,
              url: DASHBOARD_ROUTES.samples.modernWebDevelopment,
            },
            {
              title: copy.sidebar.items.samples.machineLearningFundamentals,
              url: DASHBOARD_ROUTES.samples.machineLearningFundamentals,
            },
          ],
        },
      ],
    },
    {
      title: copy.sidebar.items.studyGuides,
      url: DASHBOARD_ROUTES.studyGuides,
      icon: FileText,
      items: [
        {
          title: copy.sidebar.items.guidesCreate,
          url: DASHBOARD_ROUTES.newStudyGuide,
        },
        {
          title: copy.sidebar.items.guidesMine,
          url: DASHBOARD_ROUTES.myStudyGuides,
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
              url: DASHBOARD_ROUTES.samples.midtermReviewPsychology,
            },
            {
              title: copy.sidebar.items.samples.binaryTreesCheatSheet,
              url: DASHBOARD_ROUTES.samples.binaryTreesCheatSheet,
            },
            {
              title: copy.sidebar.items.samples.algorithmComplexityNotes,
              url: DASHBOARD_ROUTES.samples.algorithmComplexityNotes,
            },
          ],
        },
      ],
    },
    {
      title: copy.sidebar.items.resources,
      url: DASHBOARD_ROUTES.resources,
      icon: Library,
      items: [
        {
          title: copy.sidebar.items.resourcesUpload,
          url: DASHBOARD_ROUTES.uploadResource,
        },
        {
          title: copy.sidebar.items.resourcesView,
          url: DASHBOARD_ROUTES.resources,
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
              url: DASHBOARD_ROUTES.samples.neuralNetworksPaper,
            },
            {
              title: copy.sidebar.items.samples.databasesQuickReference,
              url: DASHBOARD_ROUTES.samples.databasesQuickReference,
            },
            {
              title: copy.sidebar.items.samples.cloudComputingNotes,
              url: DASHBOARD_ROUTES.samples.cloudComputingNotes,
            },
          ],
        },
      ],
    },
  ];
}

const defaultUser: React.ComponentProps<typeof NavUser>["user"] = {
  name: "Demo User",
  email: "demo@example.com",
  avatar: "/avatars/shadcn.jpg",
};

export function AppSidebar({
  user = defaultUser,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  user?: React.ComponentProps<typeof NavUser>["user"];
}) {
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
        <NavUser user={user} labels={copy.userMenu} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
