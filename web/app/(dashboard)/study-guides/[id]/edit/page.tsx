/**
 * Study guide edit route (ASK-195).
 *
 * Server Component. Fetches the guide, then hands off to the
 * client `<EditStudyGuideForm>` which wraps the shared
 * `<StudyGuideForm mode="edit">`. A 404 from `getStudyGuide`
 * routes to the dashboard not-found boundary; 403 (non-author)
 * bubbles to the dashboard error boundary.
 */
import { notFound } from "next/navigation";

import { getStudyGuide } from "@/lib/api";
import { ApiError } from "@/lib/api/errors";

import { EditStudyGuideForm } from "./edit-study-guide-form";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function EditStudyGuidePage({ params }: PageProps) {
  const { id } = await params;
  const guide = await getStudyGuideOr404(id);
  return <EditStudyGuideForm guide={guide} />;
}

async function getStudyGuideOr404(id: string) {
  try {
    return await getStudyGuide(id);
  } catch (err) {
    if (err instanceof ApiError && (err.status === 404 || err.status === 400)) {
      notFound();
    }
    throw err;
  }
}
