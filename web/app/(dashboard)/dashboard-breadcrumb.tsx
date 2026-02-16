"use client";

import { usePathname } from "next/navigation";
import { useDashboardCommonCopy } from "@/lib/features/dashboard/i18n/common/common-copy-provider";
import { DASHBOARD_ROUTES } from "@/lib/features/dashboard/navigation/routes";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";

function segmentToLabel(segment: string) {
  return segment
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function getBreadcrumbLabel(
  pathname: string,
  labels: Record<string, string>,
  homeFallback: string,
) {
  const matchedCopy = labels[pathname];

  if (matchedCopy) {
    return matchedCopy;
  }

  const segments = pathname.split("/").filter(Boolean);

  if (!segments.length) {
    return homeFallback;
  }

  return segmentToLabel(segments[segments.length - 1]);
}

export function DashboardBreadcrumb() {
  const copy = useDashboardCommonCopy();
  const pathname = usePathname();

  const breadcrumbLabelByPath: Record<string, string> = {
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

  const label = getBreadcrumbLabel(
    pathname,
    breadcrumbLabelByPath,
    copy.breadcrumb.home,
  );

  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbPage>{label}</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  );
}
