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
import { type KeyboardEvent, type ReactNode } from "react";

import type { FileResponse } from "@/lib/api/types";
import { cn, formatBytes, formatRelativeDate } from "@/lib/utils";

interface FileCardProps {
  file: FileResponse;
  variant: "list" | "grid";
  onOpen?: (file: FileResponse) => void;
  /** Slot for an inline row menu (e.g. download/rename/delete). Only shown in `list` variant. */
  rowMenu?: ReactNode;
  /** Slot for a favorite affordance. Rendered top-right on grid, inline on list. */
  favoriteButton?: ReactNode;
}

export function FileCard({
  file,
  variant,
  onOpen,
  rowMenu,
  favoriteButton,
}: FileCardProps) {
  const name = file.name === "" ? "Untitled" : file.name;
  const Icon = resolveIcon(file.mime_type);
  const isPending = file.status === "pending";
  const isClickable = Boolean(onOpen);

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
        <p className="truncate text-sm font-medium" title={name}>
          {name}
        </p>
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

function shortMimeLabel(mime: string): string {
  const subtype = mime.split("/")[1] ?? mime;
  // Strip vendor prefixes ("vnd.openxmlformats-...wordprocessingml.document" -> "word")
  const cleaned = subtype.replace(/^vnd\..+?(\w+)$/, "$1");
  return cleaned.slice(0, 6);
}

function stopPropagation(event: React.MouseEvent) {
  event.stopPropagation();
}

function stopKeyboardPropagation(event: React.KeyboardEvent) {
  event.stopPropagation();
}
