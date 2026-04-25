/**
 * Create-a-study-guide route (ASK-191).
 *
 * Server Component. Pre-fetches the caller's enrollments so the
 * client form can render a course selector populated from real
 * data. The `?course=<id>` query param is the hint we send from
 * the course detail page's "+ New guide" CTA -- when present and
 * the user is enrolled, the selector starts on that course.
 */
import { listMyEnrollments } from "@/lib/api";

import { NewStudyGuideForm } from "./new-study-guide-form";

interface PageProps {
  searchParams: Promise<{ course?: string | string[] }>;
}

function pickCourseId(value: string | string[] | undefined): string | null {
  if (!value) return null;
  return Array.isArray(value) ? (value[0] ?? null) : value;
}

export default async function NewStudyGuidePage({ searchParams }: PageProps) {
  const [{ enrollments }, params] = await Promise.all([
    listMyEnrollments(),
    searchParams,
  ]);

  return (
    <NewStudyGuideForm
      enrollments={enrollments}
      defaultCourseId={pickCourseId(params.course)}
    />
  );
}
