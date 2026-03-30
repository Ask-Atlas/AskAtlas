"use client";

import { Users, BookOpen, GraduationCap, Globe, Trash2 } from "lucide-react";
import { useState } from "react";

/**
 * File Permissions Manager Component
 *
 * Manages file_grants table entries for a specific file.
 * Allows granting view/share/delete permissions to:
 * - Individual users
 * - Courses
 * - Study guides
 * - Public access (grantee_id: 00000000-0000-0000-0000-000000000000)
 */

interface FileGrant {
  id: string; // file_grants.id
  file_id: string;
  grantee_type: "user" | "course" | "study_guide";
  grantee_id: string;
  permission: "view" | "share" | "delete";
  granted_by: string;
  created_at: string;
  // Joined data for display
  grantee_name?: string;
}

interface FilePermissionsProps {
  fileId: string;
  fileName: string;
  onClose?: () => void;
}

export function FilePermissions({ fileId, fileName, onClose }: FilePermissionsProps) {
  // In production, fetch with: GET /api/files/{fileId}/grants (you'll need to create this endpoint)
  const [grants, setGrants] = useState<FileGrant[]>([
    {
      id: crypto.randomUUID(),
      file_id: fileId,
      grantee_type: "study_guide",
      grantee_id: "sg-1",
      permission: "view",
      granted_by: "current-user-id",
      created_at: new Date().toISOString(),
      grantee_name: "Introduction to Biology",
    },
    {
      id: crypto.randomUUID(),
      file_id: fileId,
      grantee_type: "user",
      grantee_id: "user-2",
      permission: "share",
      granted_by: "current-user-id",
      created_at: new Date().toISOString(),
      grantee_name: "John Doe",
    },
  ]);

  const [isAddingGrant, setIsAddingGrant] = useState(false);
  const [newGrantType, setNewGrantType] = useState<"user" | "course" | "study_guide">("user");
  const [newGrantPermission, setNewGrantPermission] = useState<"view" | "share" | "delete">("view");

  const handleAddGrant = async () => {
    // TODO: Implement grant creation
    // POST /api/files/{fileId}/grant
    // Body: {
    //   grantee_type: newGrantType,
    //   grantee_id: selected_grantee_id,
    //   permission: newGrantPermission,
    //   granted_by: current_user_id
    // }

    console.log("Adding grant:", { fileId, newGrantType, newGrantPermission });
    setIsAddingGrant(false);
  };

  const handleRemoveGrant = async (grantId: string) => {
    // TODO: Implement grant removal
    // DELETE /api/files/{fileId}/grant
    // Body: { grant_id: grantId }
    // OR: DELETE /api/files/{fileId}/grant?grant_id={grantId}

    setGrants((prev) => prev.filter((g) => g.id !== grantId));
    console.log("Removing grant:", grantId);
  };

  const getGrantIcon = (type: string) => {
    switch (type) {
      case "user":
        return <Users className="w-4 h-4" />;
      case "course":
        return <GraduationCap className="w-4 h-4" />;
      case "study_guide":
        return <BookOpen className="w-4 h-4" />;
      default:
        return <Globe className="w-4 h-4" />;
    }
  };

  const getPermissionColor = (permission: string) => {
    switch (permission) {
      case "view":
        return "bg-green-500/10 text-green-400 border-green-500/20";
      case "share":
        return "bg-blue-500/10 text-blue-400 border-blue-500/20";
      case "delete":
        return "bg-red-500/10 text-red-400 border-red-500/20";
      default:
        return "bg-zinc-500/10 text-zinc-400 border-zinc-500/20";
    }
  };

  return (
    <div className="bg-zinc-900/40 rounded-2xl border border-zinc-800/50 p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold text-white">File Permissions</h3>
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

      {/* Current Grants */}
      <div className="space-y-3 mb-6">
        <h4 className="text-sm font-semibold uppercase tracking-wider text-zinc-500">
          Current Grants ({grants.length})
        </h4>
        {grants.length === 0 ? (
          <p className="text-sm text-zinc-500 py-4">No permissions granted yet.</p>
        ) : (
          <div className="space-y-2">
            {grants.map((grant) => (
              <div
                key={grant.id}
                className="flex items-center justify-between p-3 rounded-lg bg-[#252525] border border-zinc-800/50"
              >
                <div className="flex items-center gap-3">
                  <div className="text-zinc-400">{getGrantIcon(grant.grantee_type)}</div>
                  <div>
                    <p className="text-sm font-medium text-white">
                      {grant.grantee_name || grant.grantee_id}
                    </p>
                    <p className="text-xs text-zinc-500 capitalize">{grant.grantee_type}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={`px-2 py-1 text-xs rounded-full border ${getPermissionColor(
                      grant.permission
                    )}`}
                  >
                    {grant.permission}
                  </span>
                  <button
                    onClick={() => handleRemoveGrant(grant.id)}
                    className="p-1 hover:bg-red-500/10 rounded transition-colors"
                  >
                    <Trash2 className="w-4 h-4 text-red-500" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add New Grant */}
      {!isAddingGrant ? (
        <button
          onClick={() => setIsAddingGrant(true)}
          className="w-full px-4 py-2 rounded-lg bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 border border-orange-500/20 hover:border-orange-500/40 transition-all font-medium"
        >
          + Add Permission
        </button>
      ) : (
        <div className="p-4 rounded-lg bg-[#252525] border border-zinc-800/50 space-y-3">
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-2">
              Grant Type
            </label>
            <select
              value={newGrantType}
              onChange={(e) => setNewGrantType(e.target.value as any)}
              className="w-full px-3 py-2 rounded-lg bg-zinc-800 border border-zinc-700 text-white"
            >
              <option value="user">User</option>
              <option value="course">Course</option>
              <option value="study_guide">Study Guide</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-2">
              Permission Level
            </label>
            <select
              value={newGrantPermission}
              onChange={(e) => setNewGrantPermission(e.target.value as any)}
              className="w-full px-3 py-2 rounded-lg bg-zinc-800 border border-zinc-700 text-white"
            >
              <option value="view">View</option>
              <option value="share">Share</option>
              <option value="delete">Delete</option>
            </select>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleAddGrant}
              className="flex-1 px-4 py-2 rounded-lg bg-orange-500 hover:bg-orange-600 text-white font-medium transition-colors"
            >
              Grant Access
            </button>
            <button
              onClick={() => setIsAddingGrant(false)}
              className="flex-1 px-4 py-2 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-white font-medium transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Info */}
      <div className="mt-6 p-3 rounded-lg bg-blue-500/5 border border-blue-500/20">
        <p className="text-xs text-blue-300">
          <strong>Note:</strong> For public access, use grantee_id:{" "}
          <code className="text-blue-400">00000000-0000-0000-0000-000000000000</code>
        </p>
      </div>
    </div>
  );
}
