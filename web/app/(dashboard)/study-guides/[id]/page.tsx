import { notFound } from "next/navigation";

import { getStudyGuide, listStudyGuideGrants } from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import { StudyGuideView } from "@/lib/features/dashboard/study-guides/study-guide-view";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function StudyGuidePage({ params }: PageProps) {
  const { id } = await params;

  // listStudyGuideGrants 200s only for the owner + share/delete
  // grantees; plain viewers 403. Use that as the edit gate since
  // the detail response doesn't expose caller permissions.
  const [guide, canEdit] = await Promise.all([
    getStudyGuideOr404(id),
    listStudyGuideGrants(id)
      .then(() => true)
      .catch(() => false),
  ]);

  return <StudyGuideView guide={guide} canEdit={canEdit} />;
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
