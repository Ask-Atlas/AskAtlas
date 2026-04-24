"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { forwardRef, type ForwardedRef, useImperativeHandle } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type {
  CreateStudyGuideRequest,
  StudyGuideDetailResponse,
  UpdateStudyGuideRequest,
} from "@/lib/api/types";

// Inline error copy lives on the schema so AC3/AC4 messages match
// the ticket verbatim ("Title must be at least 3 characters" /
// "Content must be at least 10 characters").
const schema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters"),
  content: z.string().min(10, "Content must be at least 10 characters"),
  // Raw comma-separated user input; normalized to string[] on submit.
  tagsText: z.string(),
});

type FormValues = z.infer<typeof schema>;

export type StudyGuideFormField = keyof FormValues;

/**
 * Imperative handle for AC5: callers that catch an ApiError with
 * `status: "validation_error"` project its `details` onto field-level
 * messages via `setError`. The form doesn't know about ApiError --
 * the caller translates and calls this.
 */
export interface StudyGuideFormHandle {
  setError: (field: StudyGuideFormField, message: string) => void;
}

interface StudyGuideFormProps {
  mode: "create" | "edit";
  /** Required when `mode === "edit"` -- provides defaults for the fields. */
  initial?: StudyGuideDetailResponse;
  /**
   * Fires with the shaped API body. Caller handles success redirect +
   * error toast; rejections propagate so the form's `isSubmitting`
   * flag clears even on failure (see react-hook-form `handleSubmit`).
   */
  onSubmit: (
    body: CreateStudyGuideRequest | UpdateStudyGuideRequest,
  ) => Promise<void>;
  onCancel: () => void;
}

export const StudyGuideForm = forwardRef<
  StudyGuideFormHandle,
  StudyGuideFormProps
>(function StudyGuideForm(
  { mode, initial, onSubmit, onCancel },
  ref: ForwardedRef<StudyGuideFormHandle>,
) {
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    // onChange so the Save button can disable/enable reactively as the
    // user types -- matches "Submit button disabled until required
    // fields meet min length" (AC3/AC4).
    mode: "onChange",
    defaultValues: {
      title: initial?.title ?? "",
      content: initial?.content ?? "",
      tagsText: (initial?.tags ?? []).join(", "),
    },
  });

  useImperativeHandle(
    ref,
    () => ({
      setError: (field, message) => form.setError(field, { message }),
    }),
    [form],
  );

  const { isSubmitting, isValid } = form.formState;

  const handleSubmit = async (values: FormValues) => {
    const tags = values.tagsText
      .split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag.length > 0);
    await onSubmit({
      title: values.title,
      content: values.content,
      tags,
    });
  };

  const submitLabel = isSubmitting
    ? mode === "create"
      ? "Creating…"
      : "Saving…"
    : mode === "create"
      ? "Create"
      : "Save";

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Title</FormLabel>
              <FormControl>
                <Input
                  placeholder="e.g. CPTS 322 Midterm Review"
                  autoComplete="off"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="content"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Content</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="Write your study guide in markdown…"
                  rows={14}
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Markdown is supported. Rich-text editing lands later.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="tagsText"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Tags</FormLabel>
              <FormControl>
                <Input
                  placeholder="midterm, concurrency, systems-programming"
                  autoComplete="off"
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Comma-separated. Used for search and recommendations.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={!isValid || isSubmitting}>
            {submitLabel}
          </Button>
        </div>
      </form>
    </Form>
  );
});
