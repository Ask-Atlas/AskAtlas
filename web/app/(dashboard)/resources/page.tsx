"use client";

import * as React from "react";
import { useState, useEffect } from "react";
import {
  FileText,
  Download,
  MoreVertical,
  X,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  LayoutGrid,
  List as ListIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";

// Matches the API FileResponse schema
interface Document {
  id: string; // UUID
  name: string;
  size: number; // bytes
  mime_type: string;
  status: "pending" | "complete" | "failed";
  created_at: string; // ISO date
  updated_at: string; // ISO date
  favorited_at?: string | null;
  last_viewed_at?: string | null;
}

interface DocumentViewerProps {
  document: Document | null;
  onClose: () => void;
}

function DocumentViewer({ document, onClose }: DocumentViewerProps) {
  const [zoom, setZoom] = useState(1);
  const [rotation, setRotation] = useState(0);

  if (!document) return null;

  const isImage = document.mime_type.startsWith('image/');
  const isPDF = document.mime_type === 'application/pdf';
  const fileUrl = `/api/files/${document.id}`; // Assuming this serves the file content

  const handleZoomIn = () => setZoom(prev => Math.min(prev + 0.25, 3));
  const handleZoomOut = () => setZoom(prev => Math.max(prev - 0.25, 0.5));
  const handleReset = () => {
    setZoom(1);
    setRotation(0);
  };
  const handleRotate = () => setRotation(prev => (prev + 90) % 360);

  const handleDownload = () => {
    const link = document.createElement('a');
    link.href = fileUrl;
    link.download = document.name;
    link.click();
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/90 backdrop-blur-sm flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between p-4 bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800">
        <div className="flex items-center gap-3">
          <FileText className="w-5 h-5 text-zinc-500" />
          <div>
            <h2 className="font-medium text-zinc-900 dark:text-white">{document.name}</h2>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              {new Date(document.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Zoom Controls */}
          <Button variant="ghost" size="sm" onClick={handleZoomOut} disabled={zoom <= 0.5}>
            <ZoomOut className="w-4 h-4" />
          </Button>
          <span className="text-sm text-zinc-600 dark:text-zinc-400 min-w-[3rem] text-center">
            {Math.round(zoom * 100)}%
          </span>
          <Button variant="ghost" size="sm" onClick={handleZoomIn} disabled={zoom >= 3}>
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

      {/* Viewer */}
      <div className="flex-1 overflow-auto bg-zinc-100 dark:bg-zinc-800 p-4">
        <div className="max-w-full mx-auto">
          <div
            className="inline-block bg-white dark:bg-zinc-900 shadow-lg rounded-lg overflow-hidden"
            style={{
              transform: `scale(${zoom}) rotate(${rotation}deg)`,
              transformOrigin: 'top center',
              transition: 'transform 0.2s ease',
            }}
          >
            {isImage && (
              <img
                src={fileUrl}
                alt={document.name}
                className="max-w-full h-auto select-text"
                style={{ userSelect: 'text' }}
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
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState<"list" | "grid">("grid");
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null);

  useEffect(() => {
    // Fetch user's documents
    const fetchDocuments = async () => {
      try {
        const response = await fetch('/api/me/files');
        if (response.ok) {
          const data = await response.json();
          setDocuments(data.files || []);
        }
      } catch (error) {
        console.error('Failed to fetch documents:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchDocuments();
  }, []);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  const getRelativeTime = (isoString: string): string => {
    const date = new Date(isoString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);

    if (diffHours < 1) return "Just now";
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? "s" : ""} ago`;
    if (diffDays === 1) return "Yesterday";
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} week${Math.floor(diffDays / 7) > 1 ? "s" : ""} ago`;
    return date.toLocaleDateString();
  };

  const handleDownload = (document: Document) => {
    const link = document.createElement('a');
    link.href = `/api/files/${document.id}`;
    link.download = document.name;
    link.click();
  };

  if (loading) {
    return (
      <section className="space-y-4">
        <header>
          <h1 className="text-2xl font-semibold tracking-tight">My resources</h1>
        </header>
        <div className="bg-muted/50 min-h-[60vh] rounded-xl animate-pulse" />
      </section>
    );
  }

  return (
    <>
      <section className="space-y-4">
        <header>
          <h1 className="text-2xl font-semibold tracking-tight">My resources</h1>
          <p className="text-muted-foreground text-sm">
            View and manage your uploaded documents.
          </p>
        </header>

        {/* View Mode Toggle */}
        <div className="flex justify-end">
          <div className="flex items-center gap-2 bg-muted/50 p-1 rounded-lg border">
            <Button
              variant={viewMode === "list" ? "default" : "ghost"}
              size="sm"
              onClick={() => setViewMode("list")}
            >
              <ListIcon className="w-4 h-4" />
            </Button>
            <Button
              variant={viewMode === "grid" ? "default" : "ghost"}
              size="sm"
              onClick={() => setViewMode("grid")}
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Documents Display */}
        {documents.length === 0 ? (
          <div className="text-center py-16 bg-muted/50 rounded-xl">
            <FileText className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No resources found</h3>
            <p className="text-muted-foreground text-sm">Upload some documents to get started.</p>
          </div>
        ) : viewMode === "list" ? (
          <div className="space-y-2">
            {documents.map((doc) => (
              <div
                key={doc.id}
                className="flex items-center justify-between p-4 bg-card rounded-lg border hover:bg-muted/50 transition-colors group"
              >
                <button
                  onClick={() => setSelectedDocument(doc)}
                  className="flex items-center gap-4 flex-1 text-left"
                >
                  <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-orange-500/10">
                    <FileText className="w-6 h-6 text-orange-500" />
                  </div>
                  <div>
                    <h3 className="font-medium group-hover:text-orange-500 transition-colors">
                      {doc.name}
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      {formatFileSize(doc.size)} • Uploaded {getRelativeTime(doc.created_at)}
                    </p>
                  </div>
                </button>

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100 transition-opacity">
                      <MoreVertical className="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => handleDownload(doc)}>
                      <Download className="w-4 h-4 mr-2" />
                      Download
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {documents.map((doc) => (
              <div
                key={doc.id}
                className="group relative bg-card rounded-lg border hover:shadow-md transition-all overflow-hidden"
              >
                <button
                  onClick={() => setSelectedDocument(doc)}
                  className="w-full p-4 text-left"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-orange-500/10">
                      <FileText className="w-6 h-6 text-orange-500" />
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100 transition-opacity">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => handleDownload(doc)}>
                          <Download className="w-4 h-4 mr-2" />
                          Download
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>

                  <h3 className="font-medium group-hover:text-orange-500 transition-colors line-clamp-2 mb-2">
                    {doc.name}
                  </h3>

                  <p className="text-xs text-muted-foreground">
                    Uploaded {getRelativeTime(doc.created_at)}
                  </p>
                </button>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Document Viewer */}
      {selectedDocument && (
        <DocumentViewer
          document={selectedDocument}
          onClose={() => setSelectedDocument(null)}
        />
      )}
    </>
  );
}
