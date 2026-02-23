"use client";

import * as React from "react";
import { useState, useRef, useEffect } from "react";
import {
  BookOpen,
  GraduationCap,
  Flame,
  FilePlus,
  Edit3,
  FileText,
  Clock,
  MoreHorizontal,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

// ─── Types ────────────────────────────────────────────────────────────────────

interface StudyGuide {
  id: string;
  title: string;
  subject: string;
  description?: string;
  lastAccessed: string;
  progress: number;
}

interface ActivityEvent {
  id: string;
  type: "create" | "edit";
  fileName: string;
  timestamp: string;
}

// ─── Data ─────────────────────────────────────────────────────────────────────

const recentStudyGuides: StudyGuide[] = [
  {
    id: "1",
    title: "Introduction to Biology",
    subject: "Science",
    description: "Cells, DNA, ecosystems, and the foundations of life",
    lastAccessed: "2 hours ago",
    progress: 75,
  },
  {
    id: "2",
    title: "World History: WWI & WWII",
    subject: "History",
    description: "Causes, key battles, treaties, and lasting consequences",
    lastAccessed: "Yesterday",
    progress: 40,
  },
  {
    id: "3",
    title: "Algebra Fundamentals",
    subject: "Mathematics",
    description: "Variables, equations, functions, and graphing basics",
    lastAccessed: "3 days ago",
    progress: 90,
  },
  {
    id: "4",
    title: "Shakespeare's Tragedies",
    subject: "English Literature",
    description: "Hamlet, Macbeth, Othello — themes, language, and context",
    lastAccessed: "1 week ago",
    progress: 20,
  },
];

const globalActivityHistory: ActivityEvent[] = [
  { id: "g1", type: "create", fileName: "Biology Notes.pdf",        timestamp: "2 hours ago" },
  { id: "g2", type: "edit",   fileName: "Cell Structure Quiz.pdf",  timestamp: "5 hours ago" },
  { id: "g3", type: "edit",   fileName: "History Essay.docx",       timestamp: "Yesterday"   },
  { id: "g4", type: "create", fileName: "DNA Overview.pdf",         timestamp: "Yesterday"   },
  { id: "g5", type: "create", fileName: "WWI Timeline.pdf",         timestamp: "2 days ago"  },
  { id: "g6", type: "create", fileName: "Algebra Practice.pdf",     timestamp: "3 days ago"  },
  { id: "g7", type: "edit",   fileName: "WWII Notes.docx",          timestamp: "3 days ago"  },
  { id: "g8", type: "edit",   fileName: "Equations Worksheet.pdf",  timestamp: "4 days ago"  },
];

const guideActivities: Record<string, ActivityEvent[]> = {
  "1": [
    { id: "e1",   type: "create", fileName: "Biology Notes.pdf",       timestamp: "2 hours ago" },
    { id: "e1-2", type: "edit",   fileName: "Cell Structure Quiz.pdf", timestamp: "5 hours ago" },
    { id: "e1-3", type: "create", fileName: "DNA Overview.pdf",        timestamp: "Yesterday"   },
  ],
  "2": [
    { id: "e2",   type: "edit",   fileName: "History Essay.docx", timestamp: "Yesterday"  },
    { id: "e2-2", type: "create", fileName: "WWI Timeline.pdf",   timestamp: "2 days ago" },
    { id: "e2-3", type: "edit",   fileName: "WWII Notes.docx",    timestamp: "3 days ago" },
  ],
  "3": [
    { id: "e3",   type: "create", fileName: "Algebra Practice.pdf",    timestamp: "3 days ago" },
    { id: "e3-2", type: "edit",   fileName: "Equations Worksheet.pdf", timestamp: "4 days ago" },
    { id: "e3-3", type: "create", fileName: "Graphing Examples.pdf",   timestamp: "5 days ago" },
  ],
  "4": [
    { id: "e4",   type: "edit",   fileName: "Shakespeare Analysis.docx", timestamp: "1 week ago"  },
    { id: "e4-2", type: "create", fileName: "Hamlet Summary.pdf",        timestamp: "1 week ago"  },
    { id: "e4-3", type: "edit",   fileName: "Macbeth Notes.docx",        timestamp: "2 weeks ago" },
  ],
};

// ─── StatCard ─────────────────────────────────────────────────────────────────

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ElementType;
  description?: string;
  secondaryValue?: string | number;
  secondaryLabel?: string;
}

function StatCard({
  title,
  value,
  icon: Icon,
  description,
  secondaryValue,
  secondaryLabel,
}: StatCardProps) {
  return (
    <Card className="bg-[#252525] border-[#3a3a3a] flex flex-col justify-center">
      <CardHeader className="flex flex-col items-center justify-center space-y-1 pb-0 pt-3">
        <Icon className="h-4 w-4 text-orange-400" />
        <CardTitle className="text-sm font-medium text-gray-300 text-center">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center justify-center py-3">
        {secondaryValue !== undefined && secondaryLabel ? (
          <div className="flex items-center gap-6">
            <div className="flex flex-col items-center">
              <div className="text-2xl font-bold text-amber-400">{value}</div>
              {description && (
                <p className="text-xs text-gray-400 text-center mt-0.5">{description}</p>
              )}
            </div>
            <div className="h-10 w-px bg-zinc-700" />
            <div className="flex flex-col items-center">
              <div className="text-2xl font-bold text-amber-400">{secondaryValue}</div>
              <p className="text-xs text-gray-400 text-center mt-0.5">{secondaryLabel}</p>
            </div>
          </div>
        ) : (
          <>
            <div className="text-2xl font-bold text-amber-400">{value}</div>
            {description && (
              <p className="text-xs text-gray-400 text-center mt-0.5">{description}</p>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}

// ─── StudyGuideList ───────────────────────────────────────────────────────────

interface StudyGuideListProps {
  guides: StudyGuide[];
  selectedId?: string;
  onSelect?: (id: string) => void;
  onView?: (id: string) => void;
  onEdit?: (id: string) => void;
  guideActivities?: Record<string, ActivityEvent[]>;
}

function StudyGuideList({
  guides,
  selectedId,
  onSelect,
  onView,
  onEdit,
  guideActivities,
}: StudyGuideListProps) {
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpenMenuId(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className="space-y-3">
      <h2 className="text-xl font-semibold text-white">Recent Study Guides</h2>
      <div className="w-full rounded-xl border border-zinc-800 bg-zinc-900/60">
        {/* Column headers */}
        <div className="grid grid-cols-[1fr_120px_32px] gap-2 px-4 py-2 border-b border-zinc-800 bg-zinc-900">
          <span className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
            Name
          </span>
          <span className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
            Last Opened
          </span>
          <span />
        </div>

        {/* File rows */}
        <div className="divide-y divide-zinc-800/70">
          {guides.map((guide) => {
            const isSelected = selectedId === guide.id;
            const isMenuOpen = openMenuId === guide.id;
            const activities = guideActivities?.[guide.id]?.slice(0, 3) ?? [];

            return (
              <div key={guide.id}>
                <div
                  onClick={() => onSelect?.(guide.id)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => e.key === "Enter" && onSelect?.(guide.id)}
                  className={cn(
                    "group w-full grid grid-cols-[1fr_120px_32px] gap-2 items-center px-4 py-2.5 text-left transition-colors cursor-pointer",
                    isSelected
                      ? "bg-orange-500/10 text-white"
                      : "text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200",
                  )}
                >
                  {/* Name + icon + description */}
                  <div className="flex items-center gap-3 min-w-0">
                    <FileText
                      size={15}
                      className={cn(
                        "shrink-0 transition-colors",
                        isSelected
                          ? "text-orange-400"
                          : "text-zinc-500 group-hover:text-zinc-400",
                      )}
                    />
                    <div className="min-w-0">
                      <button
                        onClick={(e) => { e.stopPropagation(); onView?.(guide.id); }}
                        className={cn(
                          "block text-sm font-medium truncate text-left hover:underline underline-offset-2 transition-colors",
                          isSelected
                            ? "text-orange-300 hover:text-orange-200"
                            : "text-zinc-200 hover:text-white",
                        )}
                      >
                        {guide.title}
                      </button>
                      <span className="block text-[11px] text-zinc-500 truncate">
                        {guide.description ?? guide.subject}
                      </span>
                    </div>
                  </div>

                  {/* Last accessed */}
                  <div className="flex items-center gap-1.5 text-xs text-zinc-500">
                    <Clock size={11} className="shrink-0" />
                    <span className="truncate">{guide.lastAccessed}</span>
                  </div>

                  {/* More button + dropdown */}
                  <div
                    className="relative flex items-center justify-center"
                    ref={isMenuOpen ? menuRef : null}
                  >
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setOpenMenuId(isMenuOpen ? null : guide.id);
                      }}
                      className={cn(
                        "flex items-center justify-center w-6 h-6 rounded transition-all",
                        isMenuOpen
                          ? "opacity-100 bg-zinc-700"
                          : "opacity-0 group-hover:opacity-100 hover:bg-zinc-700",
                      )}
                    >
                      <MoreHorizontal size={14} className="text-zinc-400" />
                    </button>

                    {isMenuOpen && (
                      <div className="absolute right-0 top-7 z-50 w-32 rounded-lg border border-zinc-700 bg-zinc-800 shadow-xl overflow-hidden">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            onView?.(guide.id);
                            setOpenMenuId(null);
                          }}
                          className="w-full px-3 py-2 text-left text-sm text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors"
                        >
                          View
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            onEdit?.(guide.id);
                            setOpenMenuId(null);
                          }}
                          className="w-full px-3 py-2 text-left text-sm text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors"
                        >
                          Edit
                        </button>
                      </div>
                    )}
                  </div>
                </div>

                {/* Inline Activity Timeline — shown when selected */}
                {isSelected && activities.length > 0 && (
                  <div className="bg-zinc-900/80 px-4 py-4 border-t border-zinc-800/50">
                    <h3 className="text-[11px] font-semibold uppercase tracking-wider text-zinc-500 mb-3">
                      Recent Activity
                    </h3>
                    <div className="relative space-y-4 before:absolute before:inset-0 before:ml-[7px] before:h-full before:w-[1px] before:bg-zinc-800">
                      {activities.map((event) => (
                        <div key={event.id} className="relative flex items-start gap-3">
                          <div
                            className={cn(
                              "z-10 flex items-center justify-center w-4 h-4 rounded-full border bg-zinc-900 shrink-0",
                              event.type === "create"
                                ? "border-green-500/50"
                                : "border-amber-500/50",
                            )}
                          >
                            {event.type === "create" ? (
                              <FilePlus className="text-green-500" size={9} />
                            ) : (
                              <Edit3 className="text-amber-500" size={9} />
                            )}
                          </div>
                          <div className="flex-1 min-w-0 pt-0.5">
                            <button
                              onClick={(e) => e.stopPropagation()}
                              className="block font-medium text-xs text-zinc-200 truncate text-left hover:underline underline-offset-2 hover:text-orange-300 transition-colors"
                            >
                              {event.fileName}
                            </button>
                            <p className="text-[10px] text-zinc-500">
                              {event.type === "create" ? "Created" : "Edited"} • {event.timestamp}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function HomePage() {
  const [selectedGuideId, setSelectedGuideId] = useState<string>("1");

  return (
    <div className="dark min-h-screen bg-[#1a1a1a] p-6 text-zinc-100 font-sans">
      <div className="max-w-7xl mx-auto space-y-6">

        {/* Welcome */}
        <div className="pt-4">
          <h1 className="text-3xl font-bold text-white mb-1">Welcome back!</h1>
          <p className="text-gray-400">Here&apos;s an overview of your learning progress</p>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <StatCard
            title="Total Study Guides"
            value={24}
            icon={BookOpen}
            description="Active materials"
          />
          <StatCard
            title="Quizzes Completed"
            value={18}
            icon={GraduationCap}
            description="Problem set completions"
          />
          <StatCard
            title="Study Streak"
            value={6}
            icon={Flame}
            description="Days in a row"
            secondaryValue={12}
            secondaryLabel="Best"
          />
        </div>

        {/* Main Content */}
        <div className="flex flex-col lg:flex-row gap-8 items-start">

          {/* LEFT: Study Guide List */}
          <div className="flex-1 w-full lg:w-2/3 space-y-6">
            <StudyGuideList
              guides={recentStudyGuides}
              selectedId={selectedGuideId}
              onSelect={(id) => setSelectedGuideId(id)}
              guideActivities={guideActivities}
            />
          </div>

          {/* RIGHT: Activity History Sidebar */}
          <aside className="w-full lg:w-[350px] bg-zinc-900/40 p-6 rounded-2xl border border-zinc-800/50 shadow-xl">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-zinc-500 mb-6">
              Activity History
            </h2>
            <div className="relative space-y-8 before:absolute before:inset-0 before:ml-[11px] before:h-full before:w-[1px] before:bg-zinc-800">
              {globalActivityHistory.map((event) => (
                <div key={event.id} className="relative flex items-start gap-4">
                  <div
                    className={cn(
                      "z-10 flex items-center justify-center w-6 h-6 rounded-full border bg-[#1a1a1a] shrink-0",
                      event.type === "create"
                        ? "border-green-500/50"
                        : "border-amber-500/50",
                    )}
                  >
                    {event.type === "create" ? (
                      <FilePlus className="text-green-500" size={12} />
                    ) : (
                      <Edit3 className="text-amber-500" size={12} />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <button className="block font-medium text-sm text-zinc-100 truncate text-left hover:underline underline-offset-2 hover:text-orange-300 transition-colors">
                      {event.fileName}
                    </button>
                    <p className="text-xs text-zinc-500">
                      {event.type === "create" ? "Created" : "Edited"} • {event.timestamp}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </aside>

        </div>
      </div>
    </div>
  );
}
