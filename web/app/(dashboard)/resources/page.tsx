"use client";

import {
  Download,
  FileText,
  LayoutGrid,
  List as ListIcon,
  MoreVertical,
  RotateCcw,
  X,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import Image from "next/image";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toggleFileFavorite } from "@/lib/api";
import { API_BASE } from "@/lib/api/client";
import type { FileResponse, ListFilesResponse } from "@/lib/api/types";
import { FileCard } from "@/lib/features/dashboard/files/file-card";
import { FavoriteButton } from "@/lib/features/shared/favorite-button";
import { toast } from "@/lib/features/shared/toast/toast";

interface DocumentViewerProps {
  document: FileResponse | null;
  onClose: () => void;
}

function DocumentViewer({ document, onClose }: DocumentViewerProps) {
  const [zoom, setZoom] = useState(1);
  const [rotation, setRotation] = useState(0);

  if (!document) return null;

  const isImage = document.mime_type.startsWith("image/");
  const isPDF = document.mime_type === "application/pdf";
  // The Go API responds with a 302 to a 15-min presigned S3 URL; using
  // `/download` is the canonical path. Browsers follow 302 transparently
  // for `<img>` and `<iframe>` srcs.
  const fileUrl = `${API_BASE}/files/${document.id}/download`;

  const handleZoomIn = () => setZoom((prev) => Math.min(prev + 0.25, 3));
  const handleZoomOut = () => setZoom((prev) => Math.max(prev - 0.25, 0.5));
  const handleReset = () => {
    setZoom(1);
    setRotation(0);
  };
  const handleRotate = () => setRotation((prev) => (prev + 90) % 360);

  // The 302 lands on S3 (cross-origin), so an `<a download>` attribute
  // is silently ignored by browsers. Open in a new tab; the user gets
  // an inline view or a save dialog depending on Content-Disposition.
  const handleDownload = () => {
    window.open(fileUrl, "_blank", "noopener,noreferrer");
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/90 backdrop-blur-sm flex flex-col">
      <div className="flex items-center justify-between p-4 bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800">
        <div className="flex items-center gap-3">
          <FileText className="w-5 h-5 text-zinc-500" />
          <div>
            <h2 className="font-medium text-zinc-900 dark:text-white">
              {document.name}
            </h2>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              {new Date(document.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleZoomOut}
            disabled={zoom <= 0.5}
          >
            <ZoomOut className="w-4 h-4" />
          </Button>
          <span className="text-sm text-zinc-600 dark:text-zinc-400 min-w-[3rem] text-center">
            {Math.round(zoom * 100)}%
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleZoomIn}
            disabled={zoom >= 3}
          >
            <ZoomIn className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="sm" onClick={handleRotate}>
            <RotateCcw className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="sm" onClick={handleReset}>
            Reset
          </Button>
          <Button variant="ghost" size="sm" onClick={handleDownload}>
            <Download className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-auto bg-zinc-100 dark:bg-zinc-800 p-4">
        <div className="max-w-full mx-auto">
          <div
            className="inline-block bg-white dark:bg-zinc-900 shadow-lg rounded-lg overflow-hidden"
            style={{
              transform: `scale(${zoom}) rotate(${rotation}deg)`,
              transformOrigin: "top center",
              transition: "transform 0.2s ease",
            }}
          >
            {isImage && (
              <Image
                src={fileUrl}
                alt={document.name}
                width={800}
                height={600}
                className="max-w-full h-auto select-text"
                style={{ userSelect: "text" }}
                unoptimized
              />
            )}
            {isPDF && (
              <iframe
                src={fileUrl}
                className="w-[800px] h-[600px] border-0"
                title={document.name}
              />
            )}
            {!isImage && !isPDF && (
              <div className="p-8 text-center">
                <FileText className="w-16 h-16 text-zinc-400 mx-auto mb-4" />
                <p className="text-zinc-600 dark:text-zinc-400">
                  Preview not available for this file type.
                </p>
                <Button onClick={handleDownload} className="mt-4">
                  <Download className="w-4 h-4 mr-2" />
                  Download to view
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function ResourcesPage() {
  const [files, setFiles] = useState<FileResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState<"list" | "grid">("grid");
  const [selectedFile, setSelectedFile] = useState<FileResponse | null>(null);

  useEffect(() => {
    const fetchFiles = async () => {
      try {
        const response = await fetch("/api/me/files");
        if (response.ok) {
          const data = (await response.json()) as ListFilesResponse;
          setFiles(data.files ?? []);
        }
      } catch (error) {
        console.error("Failed to fetch files:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchFiles();
  }, []);

  const handleDownload = (file: FileResponse) => {
    // 302 → cross-origin S3, so `<a download>` is silently ignored.
    // Open in a new tab; Content-Disposition decides view vs save.
    window.open(
      `${API_BASE}/files/${file.id}/download`,
      "_blank",
      "noopener,noreferrer",
    );
  };

  const rowMenu = (file: FileResponse) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          aria-label="File actions"
          className="opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <MoreVertical className="w-4 h-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleDownload(file)}>
          <Download className="w-4 h-4 mr-2" />
          Download
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const favoriteButton = (file: FileResponse) => (
    <FavoriteButton
      initialFavorited={file.favorited_at != null}
      label={
        file.favorited_at != null
          ? "Unfavorite this file"
          : "Favorite this file"
      }
      size="sm"
      onToggle={async () => {
        try {
          const response = await toggleFileFavorite(file.id);
          setFiles((current) =>
            current.map((f) =>
              f.id === file.id
                ? { ...f, favorited_at: response.favorited_at }
                : f,
            ),
          );
          return response;
        } catch (err) {
          toast.error(err);
          throw err;
        }
      }}
    />
  );

  if (loading) {
    return (
      <section className="space-y-4">
        <header>
          <h1 className="text-2xl font-semibold tracking-tight">
            My resources
          </h1>
        </header>
        <div className="bg-muted/50 min-h-[60vh] rounded-xl animate-pulse" />
      </section>
    );
  }

  return (
    <>
      <section className="space-y-4">
        <header>
          <h1 className="text-2xl font-semibold tracking-tight">
            My resources
          </h1>
          <p className="text-muted-foreground text-sm">
            View and manage your uploaded documents.
          </p>
        </header>

        <div className="flex justify-end">
          <div className="flex items-center gap-2 bg-muted/50 p-1 rounded-lg border">
            <Button
              variant={viewMode === "list" ? "default" : "ghost"}
              size="sm"
              onClick={() => setViewMode("list")}
              aria-label="List view"
            >
              <ListIcon className="w-4 h-4" />
            </Button>
            <Button
              variant={viewMode === "grid" ? "default" : "ghost"}
              size="sm"
              onClick={() => setViewMode("grid")}
              aria-label="Grid view"
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {files.length === 0 ? (
          <div className="text-center py-16 bg-muted/50 rounded-xl">
            <FileText className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No resources found</h3>
            <p className="text-muted-foreground text-sm">
              Upload some documents to get started.
            </p>
          </div>
        ) : viewMode === "list" ? (
          <div className="space-y-2">
            {files.map((file) => (
              <FileCard
                key={file.id}
                file={file}
                variant="list"
                onOpen={setSelectedFile}
                favoriteButton={favoriteButton(file)}
                rowMenu={rowMenu(file)}
              />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {files.map((file) => (
              <FileCard
                key={file.id}
                file={file}
                variant="grid"
                onOpen={setSelectedFile}
                favoriteButton={favoriteButton(file)}
              />
            ))}
          </div>
        )}
      </section>

      {selectedFile && (
        <DocumentViewer
          document={selectedFile}
          onClose={() => setSelectedFile(null)}
        />
      )}
    </>
  );
}
