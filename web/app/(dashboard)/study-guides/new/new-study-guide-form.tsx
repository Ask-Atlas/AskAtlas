"use client";

/**
 * Client child for `/study-guides/new` (ASK-191). Picks the
 * destination course (must be one the caller is enrolled in),
 * then delegates the rest to the shared `<StudyGuideForm>`.
 */
import Link from "next/link";
import { useRouter } from "next/navigation";
import { type ReactNode, useMemo, useRef, useState } from "react";
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
      <section className="flex flex-col gap-8 py-2">
        <PageHeader />
        <div className="mx-auto w-full max-w-2xl">
          <EmptyState
            icon={<BookOpen className="size-8" aria-hidden={true} />}
            title="Join a course first"
            body="Study guides live under a course so the right people can find them. Join a section and you&rsquo;ll land back here ready to draft."
            action={
              <Button asChild>
                <Link href="/courses">Browse courses</Link>
              </Button>
            }
            className="border-border bg-muted/30 rounded-[10px] border py-12"
          />
        </div>
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
    <section className="flex flex-col gap-8 py-2">
      <PageHeader>
        <CoursePicker
          courses={courses}
          value={selectedCourseId}
          onChange={setSelectedCourseId}
        />
      </PageHeader>

      <div className="mx-auto flex w-full max-w-3xl flex-col gap-6">
        <StudyGuideForm
          ref={formRef}
          mode="create"
          onSubmit={handleSubmit}
          onCancel={handleCancel}
        />
      </div>
    </section>
  );
}

function PageHeader({ children }: { children?: ReactNode }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
      <div className="space-y-1.5">
        <h1 className="text-foreground text-[28px] font-semibold leading-tight tracking-[-0.4px]">
          New study guide
        </h1>
        <p className="text-muted-foreground text-sm">
          Draft notes, an outline, or a cheat sheet for one of your courses.
        </p>
      </div>
      {children}
    </header>
  );
}

function CoursePicker({
  courses,
  value,
  onChange,
}: {
  courses: { id: string; department: string; number: string; title: string }[];
  value: string | null;
  onChange: (id: string) => void;
}) {
  return (
    <div className="flex flex-col items-start gap-1.5 sm:items-end">
      <Label
        htmlFor="course-select"
        className="text-muted-foreground text-xs font-medium uppercase tracking-wide"
      >
        Save to
      </Label>
      <Select value={value ?? undefined} onValueChange={onChange}>
        <SelectTrigger
          id="course-select"
          className="h-9 w-full min-w-[260px] text-sm"
        >
          <SelectValue placeholder="Pick a course" />
        </SelectTrigger>
        <SelectContent align="end">
          {courses.map((course) => (
            <SelectItem key={course.id} value={course.id}>
              {course.department} {course.number} — {course.title}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
