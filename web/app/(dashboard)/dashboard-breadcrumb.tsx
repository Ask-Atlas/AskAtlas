"use client";

import { usePathname } from "next/navigation";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";

const breadcrumbLabelByPath: Record<string, string> = {
  "/home": "Home",
  "/courses": "Browse Courses",
  "/me/courses": "My Courses",
  "/study-guides": "Study Guides",
  "/study-guides/new": "Create Study Guide",
  "/me/study-guides": "My Study Guides",
  "/resources": "Resources",
  "/resources/upload": "Upload Resource",
  "/me/saved": "Starred",
};

function segmentToLabel(segment: string) {
  return segment
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function getBreadcrumbLabel(pathname: string) {
  const matchedLabel = breadcrumbLabelByPath[pathname];

  if (matchedLabel) {
    return matchedLabel;
  }

  const segments = pathname.split("/").filter(Boolean);

  if (!segments.length) {
    return "Home";
  }

  return segmentToLabel(segments[segments.length - 1]);
}

export function DashboardBreadcrumb() {
  const pathname = usePathname();
  const label = getBreadcrumbLabel(pathname);

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
