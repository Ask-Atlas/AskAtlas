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
