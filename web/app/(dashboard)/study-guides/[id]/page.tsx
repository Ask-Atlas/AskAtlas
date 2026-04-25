/**
 * Study guide reader (ASK-190).
 *
 * Server Component. Fetches the guide + caller's Clerk userId in
 * parallel, then hands off to the client `<StudyGuideView>`. A 404
 * from `getStudyGuide` triggers `notFound()` so the dashboard
 * not-found boundary handles it; everything else bubbles to the
 * dashboard error boundary.
 */
import { auth } from "@clerk/nextjs/server";
import { notFound } from "next/navigation";

import { getStudyGuide } from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import { StudyGuideView } from "@/lib/features/dashboard/study-guides/study-guide-view";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function StudyGuidePage({ params }: PageProps) {
  const { id } = await params;

  const [guide, session] = await Promise.all([getStudyGuideOr404(id), auth()]);

  const isAuthor = Boolean(
    session.userId && guide.creator.id === session.userId,
  );

  return <StudyGuideView guide={guide} isAuthor={isAuthor} />;
}

async function getStudyGuideOr404(id: string) {
  try {
    return await getStudyGuide(id);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      notFound();
    }
    throw err;
  }
}
