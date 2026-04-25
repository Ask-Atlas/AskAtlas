/**
 * Study guide reader (ASK-190).
 *
 * Server Component. Fetches the guide + caller's Clerk userId in
 * parallel, then hands off to the client `<StudyGuideView>`. A 404
 * from `getStudyGuide` triggers `notFound()` so the dashboard
 * not-found boundary handles it; everything else bubbles to the
 * dashboard error boundary.
 */
import { notFound } from "next/navigation";

import { getStudyGuide, listStudyGuideGrants } from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import { StudyGuideView } from "@/lib/features/dashboard/study-guides/study-guide-view";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function StudyGuidePage({ params }: PageProps) {
  const { id } = await params;

  // We can't compare `guide.creator.id` (internal user UUID) to
  // Clerk's `userId` (e.g. `user_xxx`) directly -- the detail
  // response doesn't expose a `caller_can_edit` flag. The
  // pragmatic gate is `listStudyGuideGrants` (ASK-211): only the
  // owner and share/delete grantees can list grants, plain
  // viewers get 403. Treating a successful response as
  // "edit-level access" covers ownership AND grantees in one
  // call. Wrap in try/catch so any error (403, network) just
  // hides the action.
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
    // 400 (malformed UUID in the path) and 404 (no row with that id)
    // both mean "this URL doesn't reference a real guide" -- route
    // both to the not-found boundary instead of the error boundary
    // so a typo in the URL doesn't look like a server crash.
    if (
      err instanceof ApiError &&
      (err.status === 404 || err.status === 400)
    ) {
      notFound();
    }
    throw err;
  }
}
