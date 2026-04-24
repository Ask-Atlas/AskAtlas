"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { X } from "lucide-react";
import {
  forwardRef,
  type ForwardedRef,
  type KeyboardEvent,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Badge } from "@/components/ui/badge";
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

import { ContentEditor } from "./content-editor";
import type {
  CreateStudyGuideRequest,
  StudyGuideDetailResponse,
  UpdateStudyGuideRequest,
} from "@/lib/api/types";
import { cn } from "@/lib/utils";

// Inline error copy lives on the schema so AC3/AC4 messages match
// the ticket verbatim ("Title must be at least 3 characters" /
// "Content must be at least 10 characters").
const schema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters"),
  content: z.string().min(10, "Content must be at least 10 characters"),
  tags: z.array(z.string()),
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
      tags: initial?.tags ?? [],
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
    await onSubmit({
      title: values.title,
      content: values.content,
      tags: values.tags,
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
                <ContentEditor
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                  placeholder="Write your study guide in markdown… Paste an internal URL to embed a live card."
                  rows={14}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                Markdown + GFM supported. Paste a study-guide / quiz / file /
                course URL to embed a live reference card. Flip to Preview to
                see the rendered article.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="tags"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Tags</FormLabel>
              <FormControl>
                <TagChipsInput
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                Type a tag and press Enter to add. Backspace on an empty input
                removes the last tag. Used for search and recommendations.
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

interface TagChipsInputProps {
  value: string[];
  onChange: (next: string[]) => void;
  disabled?: boolean;
  className?: string;
}

/**
 * Minimal tag chip input: type + Enter adds a chip, Backspace on an
 * empty input removes the last chip, click X on a chip removes it.
 * Values are trimmed + lowercased + deduped client-side to mirror the
 * server's documented normalization (see CreateStudyGuideRequest docs
 * in the generated OpenAPI types).
 *
 * Autocomplete from existing tags is deliberately out of scope until
 * a `/api/tags/suggest`-style endpoint exists.
 */
function TagChipsInput({
  value,
  onChange,
  disabled = false,
  className,
}: TagChipsInputProps) {
  const [draft, setDraft] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const commit = () => {
    const next = draft.trim().toLowerCase();
    if (next === "") return;
    if (value.includes(next)) {
      setDraft("");
      return;
    }
    onChange([...value, next]);
    setDraft("");
  };

  const remove = (tag: string) => {
    onChange(value.filter((existing) => existing !== tag));
    inputRef.current?.focus();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      commit();
      return;
    }
    // Quality-of-life: backspace on empty input pops the last chip
    // (matches the GitHub / Linear tag input convention).
    if (event.key === "Backspace" && draft === "" && value.length > 0) {
      event.preventDefault();
      remove(value[value.length - 1]!);
    }
  };

  return (
    <div
      role="group"
      aria-label="Tags"
      onClick={() => inputRef.current?.focus()}
      className={cn(
        "border-input focus-within:ring-ring flex flex-wrap items-center gap-1.5 rounded-md border bg-transparent px-2 py-1.5 focus-within:ring-2",
        disabled && "cursor-not-allowed opacity-60",
        className,
      )}
    >
      {value.map((tag) => (
        <Badge
          key={tag}
          variant="secondary"
          className="gap-1 pl-2 pr-1 text-xs font-normal"
        >
          {tag}
          <button
            type="button"
            aria-label={`Remove ${tag}`}
            disabled={disabled}
            onClick={(event) => {
              event.stopPropagation();
              remove(tag);
            }}
            className="hover:bg-muted-foreground/20 focus-visible:ring-ring -mr-0.5 inline-flex size-4 items-center justify-center rounded-full focus-visible:outline-none focus-visible:ring-1 disabled:opacity-60"
          >
            <X className="size-3" aria-hidden />
          </button>
        </Badge>
      ))}
      <input
        ref={inputRef}
        type="text"
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={commit}
        disabled={disabled}
        placeholder={
          value.length === 0 ? "e.g. midterm, concurrency" : undefined
        }
        aria-label="Add a tag"
        className="placeholder:text-muted-foreground min-w-[120px] flex-1 bg-transparent py-0.5 text-sm outline-none disabled:cursor-not-allowed"
      />
    </div>
  );
}
