"use client";

import { usePathname } from "next/navigation";
import { useDashboardCommonCopy } from "@/lib/features/dashboard/i18n/common/common-copy-provider";

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
    "/home": copy.breadcrumb.home,
    "/courses": copy.breadcrumb.browseCourses,
    "/me/courses": copy.breadcrumb.myCourses,
    "/study-guides": copy.breadcrumb.studyGuides,
    "/study-guides/new": copy.breadcrumb.createStudyGuide,
    "/me/study-guides": copy.breadcrumb.myStudyGuides,
    "/resources": copy.breadcrumb.resources,
    "/resources/upload": copy.breadcrumb.uploadResource,
    "/me/saved": copy.breadcrumb.starred,
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
