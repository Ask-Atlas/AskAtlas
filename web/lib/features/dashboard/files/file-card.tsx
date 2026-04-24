"use client";

import {
  FileArchive,
  FileAudio,
  FileImage,
  FileSpreadsheet,
  FileText,
  FileType,
  FileVideo,
  type LucideIcon,
} from "lucide-react";
import {
  type KeyboardEvent,
  type ReactNode,
  useEffect,
  useRef,
  useTransition,
} from "react";

import { Input } from "@/components/ui/input";
import type { FileResponse } from "@/lib/api/types";
import { cn, formatBytes, formatRelativeDate } from "@/lib/utils";

/**
 * When provided, the filename text is replaced with an inline rename
 * input (list variant only). The caller tracks "which file is in
 * rename mode" and commits / cancels via these callbacks.
 */
interface FileCardRenameControl {
  /**
   * Handles the rename request. Rejections are caught internally so
   * the input always closes on settle -- callers wrap their own
   * onCommit with a try/catch + toast (and may re-throw so the
   * primitive sees the rejection and closes cleanly). The caller is
   * also responsible for updating the parent state that drives
   * `rename` back to `undefined` once the request resolves.
   */
  onCommit: (newName: string) => Promise<void>;
  onCancel: () => void;
}

interface FileCardProps {
  file: FileResponse;
  variant: "list" | "grid";
  onOpen?: (file: FileResponse) => void;
  /** Slot for an inline row menu (e.g. download/rename/delete). Only shown in `list` variant. */
  rowMenu?: ReactNode;
  /** Slot for a favorite affordance. Rendered top-right on grid, inline on list. */
  favoriteButton?: ReactNode;
  /**
   * When set, the list-variant filename renders as an editable input
   * instead of text. Auto-focused, pre-selected, Enter commits,
   * Esc cancels. Ignored on grid variant -- rename UI only ships with
   * the list row menu.
   */
  rename?: FileCardRenameControl;
}

export function FileCard({
  file,
  variant,
  onOpen,
  rowMenu,
  favoriteButton,
  rename,
}: FileCardProps) {
  const name = file.name === "" ? "Untitled" : file.name;
  const Icon = resolveIcon(file.mime_type);
  const isPending = file.status === "pending";
  const isRenaming = variant === "list" && rename !== undefined;
  // Disable card-level navigation while the row is in rename mode so
  // accidental row clicks/keypresses don't fire onOpen and unmount the
  // editing state mid-edit.
  const isClickable = Boolean(onOpen) && !isRenaming;

  const fire = () => onOpen?.(file);
  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (!isClickable) return;
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      fire();
    }
  };

  return variant === "list" ? (
    <div
      role={isClickable ? "button" : undefined}
      tabIndex={isClickable ? 0 : undefined}
      onClick={isClickable ? fire : undefined}
      onKeyDown={handleKeyDown}
      className={cn(
        "group bg-card flex items-center gap-3 rounded-lg border p-3 transition-colors",
        isClickable && "hover:bg-muted/50 cursor-pointer",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
      )}
    >
      <IconTile Icon={Icon} />
      <div className="min-w-0 flex-1">
        {isRenaming && rename ? (
          <RenameInput initial={file.name} rename={rename} />
        ) : (
          <p className="truncate text-sm font-medium" title={name}>
            {name}
          </p>
        )}
        <p className="text-muted-foreground truncate text-xs">
          {isPending ? "Processing…" : formatBytes(file.size)}
          <span className="mx-1.5">·</span>
          <span className="uppercase">{shortMimeLabel(file.mime_type)}</span>
          {file.last_viewed_at && (
            <>
              <span className="mx-1.5">·</span>
              <span>Viewed {formatRelativeDate(file.last_viewed_at)}</span>
            </>
          )}
        </p>
      </div>
      {favoriteButton && (
        <div
          className="shrink-0"
          onClick={stopPropagation}
          onKeyDown={stopKeyboardPropagation}
        >
          {favoriteButton}
        </div>
      )}
      {rowMenu && (
        <div
          className="shrink-0"
          onClick={stopPropagation}
          onKeyDown={stopKeyboardPropagation}
        >
          {rowMenu}
        </div>
      )}
    </div>
  ) : (
    <div
      role={isClickable ? "button" : undefined}
      tabIndex={isClickable ? 0 : undefined}
      onClick={isClickable ? fire : undefined}
      onKeyDown={handleKeyDown}
      className={cn(
        "group bg-card relative flex flex-col gap-3 rounded-xl border p-4 transition-all",
        isClickable && "hover:shadow-md cursor-pointer",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
      )}
    >
      {favoriteButton && (
        <div
          className="absolute right-2 top-2"
          onClick={stopPropagation}
          onKeyDown={stopKeyboardPropagation}
        >
          {favoriteButton}
        </div>
      )}
      <IconTile Icon={Icon} size="lg" />
      <div className="min-w-0">
        <p
          className="line-clamp-2 text-sm font-medium leading-snug"
          title={name}
        >
          {name}
        </p>
        <p className="text-muted-foreground mt-1 text-xs">
          {isPending ? "Processing…" : formatBytes(file.size)}
        </p>
        {file.last_viewed_at && (
          <p className="text-muted-foreground mt-1 text-xs">
            Viewed {formatRelativeDate(file.last_viewed_at)}
          </p>
        )}
      </div>
    </div>
  );
}

function RenameInput({
  initial,
  rename,
}: {
  initial: string;
  rename: FileCardRenameControl;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isPending, startTransition] = useTransition();

  // Auto-focus with the filename pre-selected (filename minus extension
  // if we wanted to get fancy, but we pre-select the whole string per
  // the spec so the user can type to overwrite immediately).
  useEffect(() => {
    const el = inputRef.current;
    if (!el) return;
    el.focus();
    el.select();
  }, []);

  const commit = () => {
    const raw = inputRef.current?.value ?? "";
    const trimmed = raw.trim();
    // Empty / whitespace-only keeps the input open so the user can
    // recover without re-opening the dropdown.
    if (trimmed === "") return;
    // Identical to current closes the input without a wasted API call.
    if (trimmed === initial) {
      rename.onCancel();
      return;
    }
    startTransition(async () => {
      try {
        await rename.onCommit(trimmed);
      } catch {
        // Caller toasts + re-throws; we just let useTransition settle
        // and the caller is expected to clear rename back to undefined,
        // which unmounts this input and reveals the original filename.
        rename.onCancel();
      }
    });
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    // Prevent the outer row keyboard handler (Enter/Space → onOpen)
    // from firing while editing -- isClickable is already false during
    // rename, but this also stops the FavoriteButton/row menu wrappers
    // from intercepting the event.
    event.stopPropagation();
    if (event.key === "Enter") {
      event.preventDefault();
      commit();
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      rename.onCancel();
    }
  };

  return (
    <Input
      ref={inputRef}
      defaultValue={initial}
      disabled={isPending}
      onKeyDown={handleKeyDown}
      onClick={stopPropagation}
      aria-label="New file name"
      className="h-8 min-w-0"
    />
  );
}

function IconTile({
  Icon,
  size = "md",
}: {
  Icon: LucideIcon;
  size?: "md" | "lg";
}) {
  return (
    <div
      className={cn(
        "bg-primary/10 text-primary flex shrink-0 items-center justify-center rounded-lg",
        size === "lg" ? "size-12" : "size-10",
      )}
    >
      <Icon className={size === "lg" ? "size-6" : "size-5"} />
    </div>
  );
}

const ICON_BY_MIME: Record<string, LucideIcon> = {
  "application/pdf": FileText,
  "application/msword": FileText,
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
    FileText,
  "application/vnd.ms-excel": FileSpreadsheet,
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
    FileSpreadsheet,
  "application/vnd.ms-powerpoint": FileType,
  "application/vnd.openxmlformats-officedocument.presentationml.presentation":
    FileType,
  "application/zip": FileArchive,
  "application/epub+zip": FileArchive,
  "text/plain": FileText,
  "text/csv": FileSpreadsheet,
};

function resolveIcon(mime: string): LucideIcon {
  if (ICON_BY_MIME[mime]) return ICON_BY_MIME[mime];
  if (mime.startsWith("image/")) return FileImage;
  if (mime.startsWith("video/")) return FileVideo;
  if (mime.startsWith("audio/")) return FileAudio;
  return FileText;
}

const SHORT_LABEL_BY_MIME: Record<string, string> = {
  "application/pdf": "PDF",
  "application/msword": "DOC",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
    "DOCX",
  "application/vnd.ms-excel": "XLS",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "XLSX",
  "application/vnd.ms-powerpoint": "PPT",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation":
    "PPTX",
  "application/zip": "ZIP",
  "application/epub+zip": "EPUB",
  "text/plain": "TXT",
  "text/csv": "CSV",
};

function shortMimeLabel(mime: string): string {
  if (SHORT_LABEL_BY_MIME[mime]) return SHORT_LABEL_BY_MIME[mime];
  if (mime.startsWith("image/")) return "IMG";
  if (mime.startsWith("video/")) return "VIDEO";
  if (mime.startsWith("audio/")) return "AUDIO";
  // Unknown mime: show just the subtype (e.g. "application/unknown" -> "UNKNOW").
  const subtype = mime.split("/")[1] ?? mime;
  return subtype.slice(0, 6).toUpperCase();
}

function stopPropagation(event: React.MouseEvent) {
  event.stopPropagation();
}

function stopKeyboardPropagation(event: React.KeyboardEvent) {
  event.stopPropagation();
}
