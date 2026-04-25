"use client";

/**
 * Client child for `/study-guides/new` (ASK-191). Picks the
 * destination course (must be one the caller is enrolled in),
 * then delegates the rest to the shared `<StudyGuideForm>`.
 */
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useRef, useState } from "react";
import { BookOpen } from "lucide-react";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { createStudyGuideForCourse } from "@/lib/api";
import { ApiError } from "@/lib/api/errors";
import type {
  CreateStudyGuideRequest,
  EnrollmentResponse,
  UpdateStudyGuideRequest,
} from "@/lib/api/types";
import {
  StudyGuideForm,
  type StudyGuideFormField,
  type StudyGuideFormHandle,
} from "@/lib/features/dashboard/study-guides/study-guide-form";
import { toast } from "@/lib/features/shared/toast/toast";

interface NewStudyGuideFormProps {
  enrollments: EnrollmentResponse[];
  /**
   * Course id from `?course=<id>`. Pre-selects the matching
   * enrollment when present and valid; ignored otherwise.
   */
  defaultCourseId: string | null;
}

export function NewStudyGuideForm({
  enrollments,
  defaultCourseId,
}: NewStudyGuideFormProps) {
  const router = useRouter();
  const formRef = useRef<StudyGuideFormHandle>(null);

  // De-dupe by course id -- a user can be enrolled in multiple
  // sections of the same course over different terms; the picker
  // only cares about the destination course, not the section.
  const courses = useMemo(() => {
    const seen = new Set<string>();
    const list: EnrollmentResponse["course"][] = [];
    for (const e of enrollments) {
      if (seen.has(e.course.id)) continue;
      seen.add(e.course.id);
      list.push(e.course);
    }
    return list;
  }, [enrollments]);

  const initialCourseId =
    defaultCourseId && courses.some((c) => c.id === defaultCourseId)
      ? defaultCourseId
      : (courses[0]?.id ?? null);

  const [selectedCourseId, setSelectedCourseId] = useState<string | null>(
    initialCourseId,
  );

  if (courses.length === 0) {
    return (
      <section className="mx-auto flex w-full max-w-2xl flex-col gap-6 px-6 py-10">
        <Header />
        <EmptyState
          icon={<BookOpen className="size-8" aria-hidden={true} />}
          title="Join a course first"
          body="Study guides are created under a course. Join one to start writing."
          action={
            <Button asChild>
              <Link href="/courses">Browse courses</Link>
            </Button>
          }
          className="border-border bg-muted/30 rounded-[10px] border py-12"
        />
      </section>
    );
  }

  const handleSubmit = async (
    body: CreateStudyGuideRequest | UpdateStudyGuideRequest,
  ) => {
    if (!selectedCourseId) {
      toast.error("Pick a course before saving");
      return;
    }
    try {
      const created = await createStudyGuideForCourse(
        selectedCourseId,
        body as CreateStudyGuideRequest,
      );
      router.push(`/study-guides/${created.id}`);
    } catch (err) {
      // Surface field-level validation_error details on the form when
      // the API hands them back; otherwise fall back to a toast so
      // unstructured failures still show visibly. Do NOT rethrow --
      // react-hook-form clears `isSubmitting` either way, and a
      // bubbling rejection becomes an unhandled-promise warning.
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
    router.back();
  };

  return (
    <section className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-6 py-10">
      <Header />

      <div className="flex flex-col gap-2">
        <Label htmlFor="course-select" className="text-sm font-medium">
          Course
        </Label>
        <Select
          value={selectedCourseId ?? undefined}
          onValueChange={setSelectedCourseId}
        >
          <SelectTrigger id="course-select" className="w-full">
            <SelectValue placeholder="Pick a course" />
          </SelectTrigger>
          <SelectContent>
            {courses.map((course) => (
              <SelectItem key={course.id} value={course.id}>
                {course.department} {course.number} — {course.title}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-muted-foreground text-xs">
          New guides are scoped to a course. Pick the one this guide belongs to
          — you can&rsquo;t move it later without recreating.
        </p>
      </div>

      <StudyGuideForm
        ref={formRef}
        mode="create"
        onSubmit={handleSubmit}
        onCancel={handleCancel}
      />
    </section>
  );
}

function Header() {
  return (
    <header className="space-y-1.5">
      <h1 className="text-foreground text-[28px] font-semibold leading-tight tracking-[-0.4px]">
        New study guide
      </h1>
      <p className="text-muted-foreground text-sm">
        Draft notes, an outline, or a cheat sheet for one of your courses.
      </p>
    </header>
  );
}
