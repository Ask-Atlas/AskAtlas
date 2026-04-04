"use client";

import * as React from "react";
import { useState, useCallback } from "react";
import { Upload } from "lucide-react";

interface UploadZoneProps {
  onFilesAdded: (files: File[]) => void;
}

export function UploadZone({ onFilesAdded }: UploadZoneProps) {
  const [isDragging, setIsDragging] = useState(false);

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setIsDragging(true);
    } else if (e.type === "dragleave") {
      setIsDragging(false);
    }
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragging(false);
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        onFilesAdded(Array.from(e.dataTransfer.files));
      }
    },
    [onFilesAdded],
  );

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (e.target.files && e.target.files.length > 0) {
        onFilesAdded(Array.from(e.target.files));
      }
    },
    [onFilesAdded],
  );

  return (
    <div
      className={`border-2 border-dashed rounded-2xl p-12 text-center transition-colors flex flex-col items-center justify-center min-h-[400px] ${
        isDragging
          ? "border-orange-500 bg-orange-500/5"
          : "border-zinc-800 bg-zinc-900/20 hover:border-zinc-700 hover:bg-zinc-900/40"
      }`}
      onDragEnter={handleDrag}
      onDragLeave={handleDrag}
      onDragOver={handleDrag}
      onDrop={handleDrop}
    >
      <div className="p-4 bg-orange-500/10 rounded-full mb-4">
        <Upload className="w-8 h-8 text-orange-500" />
      </div>
      <h3 className="text-xl font-semibold text-white mb-2">
        Click or drag files to upload
      </h3>
      <p className="text-zinc-400 text-sm max-w-sm mx-auto mb-8">
        Support for PDF, JPEG, PNG, and WEBP. Maximum file size 50MB per file.
      </p>

      <label className="relative cursor-pointer bg-orange-500 hover:bg-orange-600 text-white px-6 py-3 rounded-xl font-medium transition-colors shadow-lg shadow-orange-500/20">
        <span>Select Files</span>
        <input
          type="file"
          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          multiple
          accept=".pdf,.jpg,.jpeg,.png,.webp"
          onChange={handleChange}
        />
      </label>
    </div>
  );
}
