"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Tag, X } from "lucide-react";
import {
  forwardRef,
  type ForwardedRef,
  type KeyboardEvent,
  useCallback,
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
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";

import { ContentEditor } from "./content-editor";
import { GrantsManager } from "./grants-manager";
import { VisibilityChip } from "./visibility-chip";
import type {
  CreateStudyGuideRequest,
  StudyGuideDetailResponse,
  StudyGuideVisibility,
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
  visibility: z.enum(["private", "public"]),
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
      visibility:
        (initial?.visibility as StudyGuideVisibility | undefined) ?? "private",
    },
  });

  // Tracks the current grant count so the chip can show "Shared · N"
  // even before the form field changes. Only relevant in edit mode.
  const [grantCount, setGrantCount] = useState(0);
  const handleGrantCountChange = useCallback((count: number) => {
    setGrantCount(count);
  }, []);

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
      visibility: values.visibility,
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
      <form
        onSubmit={form.handleSubmit(handleSubmit)}
        className="flex flex-col"
      >
        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem className="space-y-1">
              <FormControl>
                <Input
                  placeholder="Untitled study guide"
                  autoComplete="off"
                  aria-label="Title"
                  className="!h-auto border-0 bg-transparent px-0 text-3xl font-bold leading-tight shadow-none placeholder:text-3xl placeholder:font-bold focus-visible:ring-0 dark:bg-transparent md:text-4xl md:placeholder:text-4xl"
                  {...field}
                />
              </FormControl>
              <FormMessage className="px-0" />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="content"
          render={({ field }) => (
            <FormItem className="space-y-1">
              <FormControl>
                <ContentEditor
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                  placeholder="Write your study guide — paste a study-guide / quiz / file / course URL to embed a live card."
                  disabled={isSubmitting}
                  className="[&_.ProseMirror]:!border-0 [&_.ProseMirror]:!px-0 [&_.ProseMirror]:!min-h-[20rem]"
                />
              </FormControl>
              <FormMessage className="px-0" />
            </FormItem>
          )}
        />
        <div className="mt-6 flex flex-col gap-3 border-t pt-4 md:flex-row md:items-start md:justify-between">
          <div className="flex w-full flex-col gap-3 md:max-w-xl md:flex-1 md:flex-row md:items-start">
            <FormField
              control={form.control}
              name="tags"
              render={({ field }) => (
                <FormItem className="w-full space-y-1 md:flex-1">
                  <FormControl>
                    <div className="flex items-start gap-2">
                      <Tag
                        className="text-muted-foreground mt-2 size-4 shrink-0"
                        aria-hidden
                      />
                      <TagChipsInput
                        value={field.value}
                        onChange={field.onChange}
                        disabled={isSubmitting}
                        className="flex-1"
                      />
                    </div>
                  </FormControl>
                  <FormMessage className="px-0" />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="visibility"
              render={({ field }) => (
                <FormItem className="space-y-1">
                  <FormControl>
                    <VisibilityChip
                      visibility={field.value}
                      grantCount={mode === "edit" ? grantCount : 0}
                      disabled={isSubmitting}
                    >
                      <VisibilityPopoverBody
                        mode={mode}
                        studyGuideId={initial?.id}
                        value={field.value}
                        onChange={field.onChange}
                        disabled={isSubmitting}
                        onGrantCountChange={handleGrantCountChange}
                      />
                    </VisibilityChip>
                  </FormControl>
                  <FormMessage className="px-0" />
                </FormItem>
              )}
            />
          </div>
          <div className="flex shrink-0 justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              onClick={onCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={!isValid || isSubmitting}>
              {submitLabel}
            </Button>
          </div>
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
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const commit = () => {
    const next = draft.trim().toLowerCase();
    if (next === "") {
      setDraft("");
      setEditing(false);
      return;
    }
    if (!value.includes(next)) {
      onChange([...value, next]);
    }
    setDraft("");
    // keep input focused for rapid-add flow
    requestAnimationFrame(() => inputRef.current?.focus());
  };

  const cancel = () => {
    setDraft("");
    setEditing(false);
  };

  const remove = (tag: string) => {
    onChange(value.filter((existing) => existing !== tag));
  };

  const startEditing = () => {
    if (disabled) return;
    setEditing(true);
    requestAnimationFrame(() => inputRef.current?.focus());
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      commit();
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      cancel();
      return;
    }
    if (event.key === "Backspace" && draft === "" && value.length > 0) {
      event.preventDefault();
      remove(value[value.length - 1]!);
    }
  };

  return (
    <div
      role="group"
      aria-label="Tags"
      className={cn(
        "flex flex-wrap items-center gap-1.5",
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
      {editing ? (
        <input
          ref={inputRef}
          type="text"
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={commit}
          disabled={disabled}
          placeholder="Type a tag + Enter"
          aria-label="New tag"
          className="border-input focus-visible:ring-ring placeholder:text-muted-foreground inline-flex h-7 min-w-[140px] rounded-full border bg-transparent px-3 text-xs outline-none focus-visible:ring-2"
        />
      ) : (
        <button
          type="button"
          onClick={startEditing}
          disabled={disabled}
          className="text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-ring inline-flex h-7 items-center gap-1 rounded-full border border-dashed px-3 text-xs transition-colors focus-visible:outline-none focus-visible:ring-2 disabled:cursor-not-allowed"
        >
          <span aria-hidden>+</span>
          {value.length === 0 ? "Add tags" : "Add tag"}
        </button>
      )}
    </div>
  );
}

interface VisibilityPopoverBodyProps {
  mode: "create" | "edit";
  studyGuideId: string | undefined;
  value: StudyGuideVisibility;
  onChange: (next: StudyGuideVisibility) => void;
  disabled: boolean;
  onGrantCountChange: (count: number) => void;
}

/**
 * Popover body rendered inside `VisibilityChip`: a Private/Public
 * segmented control on top, and (edit-mode only) the GrantsManager
 * below. In create mode we show a short hint explaining the guide
 * must be saved before grants can be attached.
 */
function VisibilityPopoverBody({
  mode,
  studyGuideId,
  value,
  onChange,
  disabled,
  onGrantCountChange,
}: VisibilityPopoverBodyProps) {
  const options: Array<{ id: StudyGuideVisibility; label: string }> = [
    { id: "private", label: "Private" },
    { id: "public", label: "Public" },
  ];
  return (
    <div className="space-y-3">
      <div
        role="radiogroup"
        aria-label="Visibility"
        className="bg-muted flex rounded-md p-0.5"
      >
        {options.map((option) => {
          const checked = value === option.id;
          return (
            <button
              key={option.id}
              type="button"
              role="radio"
              aria-checked={checked}
              disabled={disabled}
              onClick={() => onChange(option.id)}
              className={cn(
                "flex-1 rounded-sm px-2 py-1 text-xs transition-colors",
                "focus-visible:ring-ring focus-visible:outline-none focus-visible:ring-2",
                checked
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {option.label}
            </button>
          );
        })}
      </div>
      {mode === "edit" && studyGuideId ? (
        <GrantsManager
          studyGuideId={studyGuideId}
          onGrantCountChange={onGrantCountChange}
        />
      ) : (
        <p className="text-muted-foreground text-xs">
          Save the guide first to share with courses or people.
        </p>
      )}
    </div>
  );
}
