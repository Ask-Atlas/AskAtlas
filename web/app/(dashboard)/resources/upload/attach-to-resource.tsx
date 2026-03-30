"use client";

import { BookOpen, GraduationCap, Link2, X } from "lucide-react";
import { useState } from "react";

/**
 * Attach File to Resource Component
 *
 * Allows attaching files to study guides or courses via junction tables:
 * - study_guide_files
 * - course_files
 *
 * API Endpoints (once tables exist):
 * - POST /api/v1/study-guides/{id}/files/{file_id}
 * - POST /api/v1/courses/{id}/files/{file_id}
 * - DELETE /api/v1/study-guides/{id}/files/{file_id}
 * - DELETE /api/v1/courses/{id}/files/{file_id}
 */

interface Resource {
  id: string;
  name: string;
  type: "study_guide" | "course";
}

interface AttachToResourceProps {
  fileId: string;
  fileName: string;
  currentAttachments?: Resource[];
  onClose?: () => void;
}

export function AttachToResource({
  fileId,
  fileName,
  currentAttachments = [],
  onClose,
}: AttachToResourceProps) {
  const [attachments, setAttachments] = useState<Resource[]>(currentAttachments);
  const [isAdding, setIsAdding] = useState(false);
  const [resourceType, setResourceType] = useState<"study_guide" | "course">("study_guide");

  // Mock available resources - in production, fetch from API
  const availableStudyGuides = [
    { id: "sg-1", name: "Introduction to Biology", type: "study_guide" as const },
    { id: "sg-2", name: "World History: WWI & WWII", type: "study_guide" as const },
    { id: "sg-3", name: "Algebra Fundamentals", type: "study_guide" as const },
  ];

  const availableCourses = [
    { id: "c-1", name: "Biology 101", type: "course" as const },
    { id: "c-2", name: "Chemistry 101", type: "course" as const },
    { id: "c-3", name: "Physics 101", type: "course" as const },
  ];

  const handleAttach = async (resource: Resource) => {
    // TODO: Implement attachment
    // For study guides:
    // POST /api/v1/study-guides/{resource.id}/files/{fileId}
    //
    // For courses:
    // POST /api/v1/courses/{resource.id}/files/{fileId}

    setAttachments((prev) => [...prev, resource]);
    setIsAdding(false);
    console.log("Attaching file", fileId, "to", resource.type, resource.id);
  };

  const handleDetach = async (resourceId: string, resourceType: "study_guide" | "course") => {
    // TODO: Implement detachment
    // For study guides:
    // DELETE /api/v1/study-guides/{resourceId}/files/{fileId}
    //
    // For courses:
    // DELETE /api/v1/courses/{resourceId}/files/{fileId}

    setAttachments((prev) => prev.filter((r) => r.id !== resourceId));
    console.log("Detaching file", fileId, "from", resourceType, resourceId);
  };

  const availableResources =
    resourceType === "study_guide" ? availableStudyGuides : availableCourses;
  const unattachedResources = availableResources.filter(
    (r) => !attachments.find((a) => a.id === r.id)
  );

  return (
    <div className="bg-zinc-900/40 rounded-2xl border border-zinc-800/50 p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold text-white">Attach to Study Materials</h3>
          <p className="text-sm text-zinc-400 mt-1">{fileName}</p>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="text-zinc-400 hover:text-white transition-colors"
          >
            Close
          </button>
        )}
      </div>

      {/* Current Attachments */}
      <div className="space-y-3 mb-6">
        <h4 className="text-sm font-semibold uppercase tracking-wider text-zinc-500">
          Currently Attached ({attachments.length})
        </h4>
        {attachments.length === 0 ? (
          <p className="text-sm text-zinc-500 py-4">Not attached to any resources yet.</p>
        ) : (
          <div className="space-y-2">
            {attachments.map((resource) => (
              <div
                key={resource.id}
                className="flex items-center justify-between p-3 rounded-lg bg-[#252525] border border-zinc-800/50"
              >
                <div className="flex items-center gap-3">
                  <div className="text-zinc-400">
                    {resource.type === "study_guide" ? (
                      <BookOpen className="w-4 h-4" />
                    ) : (
                      <GraduationCap className="w-4 h-4" />
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">{resource.name}</p>
                    <p className="text-xs text-zinc-500 capitalize">
                      {resource.type.replace("_", " ")}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => handleDetach(resource.id, resource.type)}
                  className="p-1 hover:bg-red-500/10 rounded transition-colors"
                >
                  <X className="w-4 h-4 text-red-500" />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add Attachment */}
      {!isAdding ? (
        <button
          onClick={() => setIsAdding(true)}
          className="w-full px-4 py-2 rounded-lg bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 border border-orange-500/20 hover:border-orange-500/40 transition-all font-medium"
        >
          <Link2 className="w-4 h-4 inline mr-2" />
          Attach to Resource
        </button>
      ) : (
        <div className="p-4 rounded-lg bg-[#252525] border border-zinc-800/50 space-y-3">
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-2">
              Resource Type
            </label>
            <select
              value={resourceType}
              onChange={(e) => setResourceType(e.target.value as any)}
              className="w-full px-3 py-2 rounded-lg bg-zinc-800 border border-zinc-700 text-white"
            >
              <option value="study_guide">Study Guide</option>
              <option value="course">Course</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-2">
              Select {resourceType === "study_guide" ? "Study Guide" : "Course"}
            </label>
            {unattachedResources.length === 0 ? (
              <p className="text-sm text-zinc-500 py-2">
                All available resources are already attached.
              </p>
            ) : (
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {unattachedResources.map((resource) => (
                  <button
                    key={resource.id}
                    onClick={() => handleAttach(resource)}
                    className="w-full text-left p-3 rounded-lg bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 hover:border-orange-500/50 transition-all"
                  >
                    <div className="flex items-center gap-3">
                      <div className="text-zinc-400">
                        {resource.type === "study_guide" ? (
                          <BookOpen className="w-4 h-4" />
                        ) : (
                          <GraduationCap className="w-4 h-4" />
                        )}
                      </div>
                      <span className="text-sm text-white">{resource.name}</span>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          <button
            onClick={() => setIsAdding(false)}
            className="w-full px-4 py-2 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-white font-medium transition-colors"
          >
            Cancel
          </button>
        </div>
      )}

      {/* Info */}
      <div className="mt-6 p-3 rounded-lg bg-yellow-500/5 border border-yellow-500/20">
        <p className="text-xs text-yellow-300">
          <strong>Note:</strong> These features require the <code>course_files</code> and{" "}
          <code>study_guide_files</code> junction tables to be implemented.
        </p>
      </div>
    </div>
  );
}
