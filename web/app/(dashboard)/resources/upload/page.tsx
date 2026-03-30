"use client";

/**
 * Document Upload Page
 *
 * API Integration Points:
 *
 * 1. Fetch user's files:
 *    GET /api/files
 *    Optional query params: ?course_id={id} or ?study_guide_id={id}
 *
 * 2. Upload new file:
 *    - Upload file to S3 storage (get s3_key)
 *    - Calculate file checksum (MD5/SHA256)
 *    - POST /api/files
 *      Body: { user_id, s3_key, name, mime_type, size, checksum, status }
 *
 * 3. Delete file:
 *    DELETE /api/files/{id}
 *    (Cascades to file_grants, course_files, study_guide_files)
 *
 * 4. Update file metadata:
 *    PATCH /api/files/{id}
 *    Body: { name?, ... }
 *
 * 5. Manage permissions:
 *    POST /api/files/{id}/grant
 *    Body: { grantee_type, grantee_id, permission }
 *    DELETE /api/files/{id}/grant
 *
 * 6. Attach to study guide (when tables exist):
 *    POST /api/v1/study-guides/{id}/files/{file_id}
 *    DELETE /api/v1/study-guides/{id}/files/{file_id}
 *
 * 7. Attach to course (when tables exist):
 *    POST /api/v1/courses/{id}/files/{file_id}
 *    DELETE /api/v1/courses/{id}/files/{file_id}
 */

import * as React from "react";
import { useState, useCallback } from "react";
import {
  FileText,
  Upload,
  HardDrive,
  Clock,
  LayoutGrid,
  List as ListIcon,
  X,
  Download,
  AlertCircle,
  CheckCircle2,
  File as FileIcon,
} from "lucide-react";
import { StatCard } from "@/components/ui/stat-card";
import { AttachToResource } from "./attach-to-resource";
import { FilePermissions } from "./file-permissions";

interface UploadTask {
  id: string;
  file: File;
  progress: number;
  status: "uploading" | "success" | "error";
}

// Matches the files table schema with joined data
interface Document {
  id: string; // files.id (UUID)
  name: string; // files.name
  size: number; // files.size (bytes)
  mime_type: string;
  status: "pending" | "complete" | "failed";
  s3_key: string; // files.s3_key
  created_at: string; // files.created_at
  user_id: string; // files.user_id
  // Optionally include related data from joins
  study_guides?: Array<{ id: string; name: string }>; // from study_guide_files join
  courses?: Array<{ id: string; name: string }>; // from course_files join
}

// Mock data for recent documents
// In production, fetch with: GET /api/files
const recentDocuments: Document[] = [
  {
    id: crypto.randomUUID(),
    name: "Biology Chapter 3.pdf",
    size: 2516582, // 2.4 MB in bytes
    mime_type: "application/pdf",
    status: "complete",
    s3_key: "uploads/biology-chapter-3.pdf",
    created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
    user_id: "current-user-id",
    study_guides: [{ id: "sg-1", name: "Introduction to Biology" }],
  },
  {
    id: crypto.randomUUID(),
    name: "History Essay Photo.jpg",
    size: 876544, // 856 KB in bytes
    mime_type: "image/jpeg",
    status: "complete",
    s3_key: "uploads/history-essay-photo.jpg",
    created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // Yesterday
    user_id: "current-user-id",
    courses: [{ id: "c-1", name: "World History" }],
  },
  {
    id: crypto.randomUUID(),
    name: "Math Problem Set.pdf",
    size: 1258291, // 1.2 MB in bytes
    mime_type: "application/pdf",
    status: "complete",
    s3_key: "uploads/math-problem-set.pdf",
    created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(), // 3 days ago
    user_id: "current-user-id",
    study_guides: [{ id: "sg-2", name: "Algebra Fundamentals" }],
  },
  {
    id: crypto.randomUUID(),
    name: "Chemistry Lab Diagram.png",
    size: 3250585, // 3.1 MB in bytes
    mime_type: "image/png",
    status: "complete",
    s3_key: "uploads/chemistry-lab-diagram.png",
    created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(), // 1 week ago
    user_id: "current-user-id",
    courses: [{ id: "c-2", name: "Chemistry 101" }],
  },
];

export default function UploadResourcePage() {
  const [totalDocuments, setTotalDocuments] = useState(47);
  const [totalStorage, setTotalStorage] = useState(234); // MB
  const [documents, setDocuments] = useState<Document[]>(recentDocuments);

  const [uploadQueue, setUploadQueue] = useState<UploadTask[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const [viewMode, setViewMode] = useState<"list" | "grid">("list");
  const [selectedCourse, setSelectedCourse] = useState<string | "all">("all");
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null);

  // Derive courses available for filtering
  const uniqueCourses = React.useMemo(() => {
    const courses = new Set<string>();
    documents.forEach((doc) => {
      doc.courses?.forEach((c) => courses.add(c.name));
    });
    return Array.from(courses).sort();
  }, [documents]);

  const filteredDocuments = React.useMemo(() => {
    return documents.filter(
      (doc) => selectedCourse === "all" || doc.courses?.some((c) => c.name === selectedCourse)
    );
  }, [documents, selectedCourse]);

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

  const processFiles = useCallback((files: File[]) => {
    const newTasks = files.map((f) => ({
      id: crypto.randomUUID(),
      file: f,
      progress: 0,
      status: "uploading" as const,
    }));
    
    setUploadQueue((prev) => [...prev, ...newTasks]);

    // Simulate upload progress
    newTasks.forEach((task) => {
      let currentProgress = 0;
      const interval = setInterval(() => {
        currentProgress += Math.floor(Math.random() * 20) + 10;
        if (currentProgress >= 100) {
          currentProgress = 100;
          clearInterval(interval);
          setUploadQueue((prev) =>
            prev.map((t) => (t.id === task.id ? { ...t, progress: 100, status: "success" } : t))
          );
          
          // Add to documents list on success
          setTimeout(() => {
            const newDoc: Document = {
              id: crypto.randomUUID(),
              name: task.file.name,
              size: task.file.size,
              mime_type: task.file.type || "application/octet-stream",
              status: "complete",
              s3_key: `uploads/${task.file.name.replace(/\s+/g, '-').toLowerCase()}`,
              created_at: new Date().toISOString(),
              user_id: "current-user-id",
              courses: [],
              study_guides: [],
            };
            setDocuments((prev) => [newDoc, ...prev]);
            setTotalDocuments((prev) => prev + 1);
            setTotalStorage((prev) => prev + Math.round(task.file.size / (1024 * 1024)));
            
            // Remove from queue 2s after success
            setTimeout(() => {
              setUploadQueue((prev) => prev.filter((t) => t.id !== task.id));
            }, 2000);
          }, 300);
        } else {
          setUploadQueue((prev) =>
            prev.map((t) => (t.id === task.id ? { ...t, progress: currentProgress } : t))
          );
        }
      }, 300);
    });
  }, []);

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setIsDragging(true);
    } else if (e.type === "dragleave") {
      setIsDragging(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      processFiles(Array.from(e.dataTransfer.files));
    }
  }, [processFiles]);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      processFiles(Array.from(e.target.files));
    }
  }, [processFiles]);

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-[#1a1a1a] p-6 text-zinc-900 dark:text-zinc-100 font-sans relative transition-colors duration-200">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="pt-4">
          <h1 className="text-3xl font-bold text-zinc-900 dark:text-white mb-1">
            Documents Hub
          </h1>
          <p className="text-zinc-500 dark:text-gray-400">
            Upload, associate, and manage your study materials
          </p>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <StatCard
            title="Total Documents"
            value={totalDocuments}
            icon={FileText}
            description="Files uploaded"
          />
          <StatCard
            title="Storage Used"
            value={totalStorage}
            icon={HardDrive}
            description="MB of storage"
          />
        </div>

        {/* Upload Zone (Inline) */}
        <div 
          className={`border-2 border-dashed rounded-2xl p-12 text-center transition-all duration-200 flex flex-col items-center justify-center min-h-[300px] ${
            isDragging
              ? "border-orange-500 bg-orange-500/5 shadow-[0_0_30px_rgba(249,115,22,0.1)]"
              : "border-zinc-300 dark:border-zinc-800 bg-white dark:bg-zinc-900/20 hover:border-zinc-400 dark:hover:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-900/40 shadow-sm dark:shadow-none"
          }`}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
        >
          <div className="p-4 bg-orange-500/10 rounded-full mb-4 group-hover:bg-orange-500/20 transition-colors">
            <Upload className="w-8 h-8 text-orange-500" />
          </div>
          <h3 className="text-xl font-semibold text-zinc-900 dark:text-white mb-2">
            Click or drag files to upload
          </h3>
          <p className="text-zinc-500 dark:text-zinc-400 text-sm max-w-md mx-auto mb-8">
            Support for PDF, DOCX, PPT, PPTX, TXT, EPUB, JPEG, PNG, and WEBP. Maximum file size 50MB per file.
          </p>

          <label className="relative cursor-pointer bg-orange-500 hover:bg-orange-600 text-white px-8 py-3 rounded-xl font-medium transition-all shadow-lg shadow-orange-500/20 active:scale-95">
            <span>Select Files</span>
            <input
              type="file"
              className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
              multiple
              accept=".pdf,.docx,.doc,.ppt,.pptx,.txt,.epub,.jpg,.jpeg,.png,.webp"
              onChange={handleChange}
            />
          </label>
        </div>

        {/* Upload Queue */}
        {uploadQueue.length > 0 && (
          <div className="bg-white dark:bg-zinc-900/40 rounded-2xl border border-zinc-200 dark:border-zinc-800/50 p-6 space-y-4 shadow-sm dark:shadow-none">
            <h3 className="text-lg font-semibold text-zinc-900 dark:text-white flex items-center gap-2">
              <Upload className="w-5 h-5 text-zinc-500 dark:text-zinc-400" /> Active Uploads
            </h3>
            <div className="space-y-3">
              {uploadQueue.map((task) => (
                <div key={task.id} className="p-4 rounded-xl bg-zinc-50 dark:bg-[#252525] border border-zinc-200 dark:border-zinc-800/50">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-3">
                      <FileIcon className="w-5 h-5 text-zinc-500 dark:text-zinc-400" />
                      <div>
                        <p className="text-sm font-medium text-zinc-900 dark:text-white">{task.file.name}</p>
                        <p className="text-xs text-zinc-500">{formatFileSize(task.file.size)}</p>
                      </div>
                    </div>
                    {task.status === "uploading" && <span className="text-sm font-medium text-orange-500">{task.progress}%</span>}
                    {task.status === "success" && <CheckCircle2 className="w-5 h-5 text-green-500" />}
                    {task.status === "error" && <AlertCircle className="w-5 h-5 text-red-500" />}
                  </div>
                  {task.status === "uploading" && (
                    <div className="h-1.5 w-full bg-zinc-200 dark:bg-zinc-800 rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-orange-500 transition-all duration-300 ease-out" 
                        style={{ width: `${task.progress}%` }} 
                      />
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Documents Toolbar */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 py-4 mt-8">
          <div className="flex items-center gap-2 overflow-x-auto pb-2 flex-1 w-full sm:w-auto">
            <button
              onClick={() => setSelectedCourse("all")}
              className={`px-4 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors border ${
                selectedCourse === "all" 
                  ? "bg-zinc-900 text-white dark:bg-white dark:text-black border-transparent" 
                  : "bg-white text-zinc-600 border-zinc-200 hover:bg-zinc-50 dark:bg-zinc-800 dark:hover:bg-zinc-700 dark:text-zinc-300 dark:border-zinc-700"
              }`}
            >
              All Courses
            </button>
            {uniqueCourses.map((course) => (
              <button
                key={course}
                onClick={() => setSelectedCourse(course)}
                className={`px-4 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors border ${
                  selectedCourse === course 
                    ? "bg-purple-500 text-white border-transparent" 
                    : "bg-white text-zinc-600 border-zinc-200 hover:bg-zinc-50 dark:bg-zinc-800 dark:hover:bg-zinc-700 dark:text-zinc-300 dark:border-zinc-700"
                }`}
              >
                {course}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-2 bg-white dark:bg-zinc-900/50 p-1.5 rounded-lg border border-zinc-200 dark:border-zinc-800/50 shrink-0 shadow-sm dark:shadow-none">
            <button
              onClick={() => setViewMode("list")}
              className={`p-2 rounded-md transition-colors ${viewMode === "list" ? "bg-zinc-100 dark:bg-[#252525] text-zinc-900 dark:text-white shadow-sm" : "text-zinc-500 hover:text-zinc-900 dark:hover:text-white"}`}
              title="List View"
            >
              <ListIcon className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode("grid")}
              className={`p-2 rounded-md transition-colors ${viewMode === "grid" ? "bg-zinc-100 dark:bg-[#252525] text-zinc-900 dark:text-white shadow-sm" : "text-zinc-500 hover:text-zinc-900 dark:hover:text-white"}`}
              title="Grid View"
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Documents Display */}
        {filteredDocuments.length === 0 ? (
          <div className="text-center py-16 bg-white dark:bg-zinc-900/20 rounded-2xl border border-zinc-200 dark:border-zinc-800/50 shadow-sm dark:shadow-none">
            <FileText className="w-12 h-12 text-zinc-300 dark:text-zinc-700 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-zinc-900 dark:text-white mb-2">No documents found</h3>
            <p className="text-zinc-500 text-sm">Upload a document or change your course filter.</p>
          </div>
        ) : viewMode === "list" ? (
          <div className="bg-white dark:bg-zinc-900/40 rounded-2xl border border-zinc-200 dark:border-zinc-800/50 shadow-sm dark:shadow-xl overflow-hidden">
            <div className="divide-y divide-zinc-200 dark:divide-zinc-800/50">
              {filteredDocuments.map((doc) => (
                <button
                  key={doc.id}
                  onClick={() => setSelectedDocument(doc)}
                  className="w-full text-left p-4 hover:bg-zinc-50 dark:hover:bg-[#252525] transition-all group flex items-center justify-between"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-orange-500/10 shrink-0">
                      <FileText className="w-6 h-6 text-orange-500" />
                    </div>
                    <div>
                      <h3 className="font-medium text-zinc-900 dark:text-white group-hover:text-orange-500 dark:group-hover:text-orange-300 transition-colors">
                        {doc.name}
                      </h3>
                      <p className="text-sm text-zinc-500 mt-0.5">
                        {formatFileSize(doc.size)} • {getRelativeTime(doc.created_at)}
                      </p>
                    </div>
                  </div>
                  
                  <div className="hidden md:flex gap-2">
                    {doc.courses?.map((course) => (
                      <span key={course.id} className="px-2.5 py-1 text-xs rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
                        {course.name}
                      </span>
                    ))}
                    {doc.study_guides?.map((sg) => (
                      <span key={sg.id} className="px-2.5 py-1 text-xs rounded-full bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
                        {sg.name}
                      </span>
                    ))}
                  </div>
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {filteredDocuments.map((doc) => (
              <button
                key={doc.id}
                onClick={() => setSelectedDocument(doc)}
                className="w-full text-left p-5 rounded-2xl bg-white dark:bg-zinc-900/40 hover:bg-zinc-50 dark:hover:bg-[#252525] border border-zinc-200 dark:border-zinc-800/50 dark:hover:border-zinc-700 transition-all shadow-sm dark:shadow-none hover:shadow-md group relative flex flex-col h-full"
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-orange-500/10 shrink-0">
                    <FileText className="w-6 h-6 text-orange-500" />
                  </div>
                  <Download className="w-5 h-5 text-zinc-400 dark:text-zinc-600 opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
                
                <h3 className="font-medium text-zinc-900 dark:text-white group-hover:text-orange-500 dark:group-hover:text-orange-300 transition-colors line-clamp-2 mb-2 flex-grow">
                  {doc.name}
                </h3>
                
                <p className="text-xs text-zinc-500 mb-3 block">
                  {formatFileSize(doc.size)} • {getRelativeTime(doc.created_at)}
                </p>

                <div className="flex flex-wrap gap-1 mt-auto">
                  {doc.courses?.slice(0, 2).map((course) => (
                    <span key={course.id} className="inline-block px-2 py-0.5 text-[10px] rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20 truncate max-w-full">
                      {course.name}
                    </span>
                  ))}
                  {((doc.courses?.length || 0) > 2) && (
                    <span className="inline-block px-2 py-0.5 text-[10px] rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700">
                      +{(doc.courses?.length || 0) - 2}
                    </span>
                  )}
                </div>
              </button>
            ))}
          </div>
        )}

      </div>

      {/* Document Details Drawer/Modal Overlay */}
      {selectedDocument && (
        <div className="fixed inset-0 z-50 flex justify-end bg-black/40 dark:bg-black/60 backdrop-blur-sm transition-opacity">
          {/* Dismiss background */}
          <div className="absolute inset-0 cursor-pointer" onClick={() => setSelectedDocument(null)} />
          
          {/* Side Panel */}
          <div className="relative w-full max-w-md h-full bg-white dark:bg-[#1a1a1a] border-l border-zinc-200 dark:border-zinc-800 shadow-2xl flex flex-col transform transition-transform animate-in slide-in-from-right">
            <div className="flex items-center justify-between p-6 border-b border-zinc-100 dark:border-zinc-800/50">
              <h2 className="text-xl font-semibold text-zinc-900 dark:text-white">Document Details</h2>
              <button 
                onClick={() => setSelectedDocument(null)}
                className="p-2 text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-6 space-y-6">
              {/* Top info */}
              <div className="flex gap-4 items-start">
                <div className="p-4 rounded-2xl bg-orange-500/10 flex items-center justify-center shrink-0">
                  <FileText className="w-8 h-8 text-orange-500" />
                </div>
                <div>
                  <h3 className="text-lg font-medium text-zinc-900 dark:text-white block mb-1 break-all">
                    {selectedDocument.name}
                  </h3>
                  <div className="text-sm text-zinc-500 dark:text-zinc-400 flex flex-col gap-1">
                    <span>{formatFileSize(selectedDocument.size)} • {selectedDocument.mime_type.split("/").pop()?.toUpperCase() || "FILE"}</span>
                    <span>Uploaded {getRelativeTime(selectedDocument.created_at)}</span>
                  </div>
                </div>
              </div>

              {/* Action */}
              <button 
                onClick={() => alert(`Mock downloading: ${selectedDocument.name}`)}
                className="w-full py-3 rounded-xl bg-orange-500 hover:bg-orange-600 text-white font-semibold flex items-center justify-center gap-2 transition-colors active:scale-[0.98] shadow-lg shadow-orange-500/20"
              >
                <Download className="w-5 h-5" /> Download File
              </button>

              <hr className="border-zinc-100 dark:border-zinc-800/50" />

              {/* Render User AttachToResource */}
              <AttachToResource 
                fileId={selectedDocument.id} 
                fileName={selectedDocument.name} 
                currentAttachments={[...(selectedDocument.courses?.map(c => ({...c, type: 'course' as const})) || []), ...(selectedDocument.study_guides?.map(s => ({...s, type: 'study_guide' as const})) || [])]}
              />

              {/* Render User FilePermissions */}
              <FilePermissions 
                fileId={selectedDocument.id}
                fileName={selectedDocument.name}
              />

            </div>
          </div>
        </div>
      )}

    </div>
  );
}
