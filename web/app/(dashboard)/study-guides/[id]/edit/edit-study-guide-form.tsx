"use client";

import { Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useRef, useState, useTransition } from "react";

import { Button } from "@/components/ui/button";
import {
  createStudyGuideGrant,
  deleteStudyGuide,
  listMyEnrollments,
  listStudyGuideGrants,
  revokeStudyGuideGrant,
  updateStudyGuide,
} from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import type {
  CreateStudyGuideRequest,
  StudyGuideDetailResponse,
  UpdateStudyGuideRequest,
} from "@/lib/api/types";
import type { GrantsManagerActions } from "@/lib/features/dashboard/study-guides/grants-manager";
import {
  StudyGuideForm,
  type StudyGuideFormField,
  type StudyGuideFormHandle,
} from "@/lib/features/dashboard/study-guides/study-guide-form";
import { ConfirmationDialog } from "@/lib/features/shared/confirmation-dialog";
import { toast } from "@/lib/features/shared/toast/toast";

interface EditStudyGuideFormProps {
  guide: StudyGuideDetailResponse;
}

export function EditStudyGuideForm({ guide }: EditStudyGuideFormProps) {
  const router = useRouter();
  const formRef = useRef<StudyGuideFormHandle>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [isDeleting, startDeleteTransition] = useTransition();

  const courseLabel = `${guide.course.department} ${guide.course.number}`;

  const grantActions: GrantsManagerActions = {
    listGrants: (id) => listStudyGuideGrants(id),
    listEnrollments: () => listMyEnrollments(),
    createGrant: (id, body) => createStudyGuideGrant(id, body),
    revokeGrant: (id, body) => revokeStudyGuideGrant(id, body),
  };

  const handleSubmit = async (
    body: CreateStudyGuideRequest | UpdateStudyGuideRequest,
  ) => {
    try {
      await updateStudyGuide(guide.id, body as UpdateStudyGuideRequest);
      toast.success("Study guide saved");
      router.push(`/study-guides/${guide.id}`);
    } catch (err) {
      if (err instanceof ApiError && err.body?.status === "validation_error") {
        const details = err.body.details;
        const fields: StudyGuideFormField[] = ["title", "content", "tags"];
        let projected = false;
        if (details && typeof details === "object") {
          for (const field of fields) {
            const message = (details as Record<string, unknown>)[field];
            if (typeof message === "string") {
              formRef.current?.setError(field, message);
              projected = true;
            }
          }
        }
        if (!projected) toast.error(err);
      } else {
        toast.error(err);
      }
    }
  };

  const handleCancel = () => {
    router.push(`/study-guides/${guide.id}`);
  };

  const handleConfirmDelete = () => {
    startDeleteTransition(async () => {
      try {
        await deleteStudyGuide(guide.id);
        toast.success("Study guide deleted");
        router.push(`/courses/${guide.course.id}`);
      } catch (err) {
        toast.error(err);
      } finally {
        setConfirmDeleteOpen(false);
      }
    });
  };

  return (
    <section className="mx-auto flex w-full max-w-3xl flex-col gap-8 py-2">
      <header className="flex flex-col gap-1.5">
        <p className="text-muted-foreground text-[12px] font-medium uppercase tracking-wide">
          Editing · {courseLabel}
        </p>
        <h1 className="text-foreground text-[28px] font-semibold leading-tight tracking-[-0.4px]">
          {guide.title || "Untitled study guide"}
        </h1>
      </header>

      <StudyGuideForm
        ref={formRef}
        mode="edit"
        initial={guide}
        onSubmit={handleSubmit}
        onCancel={handleCancel}
        grantActions={grantActions}
        aiEdit={{ guideId: guide.id, title: guide.title }}
      />

      <div className="border-destructive/30 bg-destructive/5 flex flex-col gap-3 rounded-[10px] border p-5">
        <div className="flex flex-col gap-1">
          <h2 className="text-foreground text-[15px] font-semibold tracking-[-0.2px]">
            Danger zone
          </h2>
          <p className="text-muted-foreground text-[13px]">
            Deleting this guide also removes its quizzes and any attached files
            / resources. This can&rsquo;t be undone.
          </p>
        </div>
        <div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="text-destructive hover:text-destructive border-destructive/30"
            onClick={() => setConfirmDeleteOpen(true)}
            disabled={isDeleting}
          >
            <Trash2 className="size-3.5" aria-hidden={true} />
            Delete study guide
          </Button>
        </div>
      </div>

      <ConfirmationDialog
        open={confirmDeleteOpen}
        onOpenChange={setConfirmDeleteOpen}
        title="Delete this study guide?"
        description="This permanently deletes the guide, its quizzes, and all attached files / resources for everyone who can see it. This can't be undone."
        confirmLabel={isDeleting ? "Deleting…" : "Delete"}
        cancelLabel="Cancel"
        destructive
        disabled={isDeleting}
        onConfirm={handleConfirmDelete}
      />
    </section>
  );
}
