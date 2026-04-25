"use client";

import { Fragment } from "react";
import { usePathname } from "next/navigation";
import { useDashboardCommonCopy } from "@/lib/features/dashboard/i18n/common/common-copy-provider";
import { DASHBOARD_ROUTES } from "@/lib/features/dashboard/navigation/routes";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

import { useDashboardBreadcrumb } from "./dashboard-breadcrumb-context";

function segmentToLabel(segment: string) {
  return segment
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function DashboardBreadcrumb() {
  const copy = useDashboardCommonCopy();
  const pathname = usePathname();
  const { currentLabel } = useDashboardBreadcrumb();

  const labelByPath: Record<string, string> = {
    [DASHBOARD_ROUTES.home]: copy.breadcrumb.home,
    [DASHBOARD_ROUTES.courses]: copy.breadcrumb.browseCourses,
    [DASHBOARD_ROUTES.myCourses]: copy.breadcrumb.myCourses,
    [DASHBOARD_ROUTES.studyGuides]: copy.breadcrumb.studyGuides,
    [DASHBOARD_ROUTES.newStudyGuide]: copy.breadcrumb.createStudyGuide,
    [DASHBOARD_ROUTES.myStudyGuides]: copy.breadcrumb.myStudyGuides,
    [DASHBOARD_ROUTES.resources]: copy.breadcrumb.resources,
    [DASHBOARD_ROUTES.uploadResource]: copy.breadcrumb.uploadResource,
    [DASHBOARD_ROUTES.saved]: copy.breadcrumb.starred,
  };

  const segments = pathname.split("/").filter(Boolean);

  if (segments.length === 0) {
    return (
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbPage>{copy.breadcrumb.home}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    );
  }

  // Build the trail. For each segment we want a label and an href.
  // The last segment renders as `BreadcrumbPage` (not a link); the
  // rest are clickable links back up the tree. Page-provided labels
  // (via context) take precedence on the trailing segment so a
  // dynamic route like `/courses/[uuid]` shows the human label
  // instead of the raw UUID.
  const items = segments.map((segment, index) => {
    const isLast = index === segments.length - 1;
    const href = "/" + segments.slice(0, index + 1).join("/");
    const labelFromCopy = labelByPath[href];
    const label = isLast
      ? (currentLabel ?? labelFromCopy ?? segmentToLabel(segment))
      : (labelFromCopy ?? segmentToLabel(segment));
    return { href, label, isLast };
  });

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {items.map((item, index) => (
          <Fragment key={item.href}>
            {index > 0 ? <BreadcrumbSeparator /> : null}
            <BreadcrumbItem>
              {item.isLast ? (
                <BreadcrumbPage>{item.label}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink href={item.href}>{item.label}</BreadcrumbLink>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
