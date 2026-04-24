"use client";

import { Download, FileText } from "lucide-react";

import { Button } from "@/components/ui/button";

import { useEntityRef } from "./entity-ref-context";
import { MissingRef } from "./study-guide-ref-card";

interface Props {
  id: string;
  inline?: boolean;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

export function FileRefCard({ id, inline }: Props) {
  const { summary, status } = useEntityRef("file", id);

  if (summary == null) {
    return <MissingRef label="File" inline={inline} status={status} />;
  }

  const downloadHref = `/api/files/${id}/download`;

  if (inline) {
    return (
      <a
        href={downloadHref}
        className="bg-primary/10 text-primary hover:bg-primary/20 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-sm font-medium no-underline"
      >
        <FileText className="size-3.5" aria-hidden />
        {summary.name ?? "File"}
      </a>
    );
  }

  return (
    <div className="my-3 flex items-center justify-between gap-3 rounded-lg border p-3">
      <div className="flex min-w-0 flex-col gap-1">
        <div className="flex items-center gap-2">
          <FileText
            className="text-muted-foreground size-4 shrink-0"
            aria-hidden
          />
          <span className="text-foreground truncate font-medium">
            {summary.name ?? "File"}
          </span>
        </div>
        <div className="text-muted-foreground flex items-center gap-3 text-xs">
          {typeof summary.size === "number" ? (
            <span>{formatBytes(summary.size)}</span>
          ) : null}
          {summary.mime_type ? <span>{summary.mime_type}</span> : null}
        </div>
      </div>
      <Button asChild size="sm" variant="outline" className="shrink-0">
        <a href={downloadHref} className="no-underline">
          <Download className="size-4" aria-hidden />
          Download
        </a>
      </Button>
    </div>
  );
}
