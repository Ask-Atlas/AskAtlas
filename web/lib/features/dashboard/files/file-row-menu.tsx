"use client";

import { Loader2, MoreVertical } from "lucide-react";
import {
  type KeyboardEvent,
  useEffect,
  useRef,
  useState,
  useTransition,
} from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import type { FileResponse } from "@/lib/api/types";
import { ConfirmationDialog } from "@/lib/features/shared/confirmation-dialog";

type Mode = "idle" | "renaming" | "confirming-delete";

interface FileRowMenuProps {
  file: FileResponse;
  /**
   * Rename the file to `newName`. Rejections are caught internally so
   * the inline input always closes; callers are expected to wrap this
   * with their own try/catch + toast (and may re-throw so the menu
   * sees the rejection and closes cleanly). The `file.name` prop stays
   * the source of truth -- on error the row re-renders with the
   * original name because we never mutate it ourselves.
   */
  onRename: (newName: string) => Promise<void>;
  /** Fires after the delete confirmation dialog is confirmed. Same error contract as `onRename`. */
  onDelete: () => Promise<void>;
}

export function FileRowMenu({ file, onRename, onDelete }: FileRowMenuProps) {
  const [mode, setMode] = useState<Mode>("idle");
  const [isPending, startTransition] = useTransition();

  if (mode === "renaming") {
    const handleCommit = (raw: string) => {
      const trimmed = raw.trim();
      // AC6: empty (or whitespace-only) keeps the input open so the
      // user can either type a real name or Esc to cancel explicitly.
      if (trimmed === "") return;
      // Edge case: identical to current closes the input without a
      // wasted API call.
      if (trimmed === file.name) {
        setMode("idle");
        return;
      }
      startTransition(async () => {
        try {
          await onRename(trimmed);
        } catch {
          // Caller toasts; we just revert to idle so the row re-renders
          // with its original filename (since we never mutated `file`).
        } finally {
          setMode("idle");
        }
      });
    };

    return (
      <RenameInput
        initial={file.name}
        pending={isPending}
        onCommit={handleCommit}
        onCancel={() => setMode("idle")}
      />
    );
  }

  const handleConfirmDelete = () => {
    startTransition(async () => {
      try {
        await onDelete();
      } catch {
        // Caller toasts.
      } finally {
        setMode("idle");
      }
    });
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            aria-label="File actions"
            disabled={isPending}
          >
            {isPending ? (
              <Loader2 className="size-4 animate-spin" aria-hidden />
            ) : (
              <MoreVertical className="size-4" aria-hidden />
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => setMode("renaming")}>
            Rename
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => setMode("confirming-delete")}
            className="text-destructive focus:text-destructive"
          >
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <ConfirmationDialog
        open={mode === "confirming-delete"}
        onOpenChange={(next) => {
          if (!next) setMode("idle");
        }}
        title="Delete file?"
        description={`Delete "${file.name}"? This can't be undone.`}
        confirmLabel={isPending ? "Deleting…" : "Delete"}
        cancelLabel="Cancel"
        destructive
        disabled={isPending}
        onConfirm={handleConfirmDelete}
      />
    </>
  );
}

function RenameInput({
  initial,
  pending,
  onCommit,
  onCancel,
}: {
  initial: string;
  pending: boolean;
  onCommit: (value: string) => void;
  onCancel: () => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);

  // AC2: the input auto-focuses with the filename pre-selected so the
  // user can start typing to overwrite without first highlighting.
  useEffect(() => {
    const el = inputRef.current;
    if (!el) return;
    el.focus();
    el.select();
  }, []);

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      onCommit(inputRef.current?.value ?? "");
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      onCancel();
    }
  };

  return (
    <Input
      ref={inputRef}
      defaultValue={initial}
      disabled={pending}
      onKeyDown={handleKeyDown}
      aria-label="New file name"
      className="h-8 min-w-0"
    />
  );
}
