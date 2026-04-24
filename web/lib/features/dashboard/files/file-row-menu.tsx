"use client";

import { Loader2, MoreVertical } from "lucide-react";
import { useState, useTransition } from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { FileResponse } from "@/lib/api/types";
import { ConfirmationDialog } from "@/lib/features/shared/confirmation-dialog";

interface FileRowMenuProps {
  file: FileResponse;
  /**
   * Signals that the user picked "Rename". The rename UX itself lives
   * on the FileCard (via its `rename` prop) so the input can appear
   * in-place of the filename -- the menu only surfaces intent. Caller
   * is responsible for setting FileCard into rename mode.
   */
  onStartRename: () => void;
  /**
   * Fires after the delete confirmation dialog is confirmed. Rejections
   * are caught internally so the dialog always closes; callers wrap
   * their own onDelete with a try/catch + toast (matching the ASK-182
   * SectionMembershipButton error contract).
   */
  onDelete: () => Promise<void>;
}

export function FileRowMenu({
  file,
  onStartRename,
  onDelete,
}: FileRowMenuProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [isPending, startTransition] = useTransition();

  const handleConfirmDelete = () => {
    startTransition(async () => {
      try {
        await onDelete();
      } catch {
        // Caller toasts.
      } finally {
        setConfirmOpen(false);
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
          <DropdownMenuItem onSelect={onStartRename}>Rename</DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => setConfirmOpen(true)}
            className="text-destructive focus:text-destructive"
          >
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <ConfirmationDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
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
