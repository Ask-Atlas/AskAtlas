"use client";

import { useState } from "react";

import { joinSection, leaveSection } from "@/lib/api";
import type { SectionSummary } from "@/lib/api/types";
import { toast } from "@/lib/features/shared/toast/toast";
import { cn } from "@/lib/utils";

import { SectionMembershipButton } from "./section-membership-button";

type Membership = "member" | "not-member" | "unknown";

interface SectionRowProps {
  courseId: string;
  section: SectionSummary;
  /**
   * Initial membership state for this section. The page determines this
   * by checking the user's enrollments — pass `"unknown"` only when the
   * caller can't pre-resolve it (the button will still render correctly).
   */
  initialMembership: Membership;
  className?: string;
}

export function SectionRow({
  courseId,
  section,
  initialMembership,
  className,
}: SectionRowProps) {
  const [membership, setMembership] = useState<Membership>(initialMembership);

  const sectionLabel = section.section_code
    ? `Section ${section.section_code}`
    : "Section";

  const handleJoin = async () => {
    try {
      await joinSection(courseId, section.id);
      setMembership("member");
    } catch (err) {
      toast.error(err);
      throw err;
    }
  };

  const handleLeave = async () => {
    try {
      await leaveSection(courseId, section.id);
      setMembership("not-member");
    } catch (err) {
      toast.error(err);
      throw err;
    }
  };

  return (
    <div
      className={cn(
        "flex items-center justify-between gap-4 px-5 py-3.5",
        className,
      )}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3.5">
        <span className="rounded-md bg-zinc-100 px-2 py-0.5 font-mono text-[11px] font-semibold tracking-[-0.2px] text-zinc-800">
          {sectionLabel}
        </span>
        <div className="flex min-w-0 flex-col gap-0.5">
          <p className="truncate text-[14px] font-medium text-zinc-950">
            {section.instructor_name ?? "Instructor TBA"}
          </p>
          <p className="truncate text-[12px] text-zinc-500">
            <span>{section.term}</span>
            <span className="px-2 text-zinc-300" aria-hidden={true}>
              ·
            </span>
            <span>
              {section.member_count}{" "}
              {section.member_count === 1 ? "member" : "members"}
            </span>
          </p>
        </div>
      </div>
      <SectionMembershipButton
        membership={membership}
        onJoin={handleJoin}
        onLeave={handleLeave}
      />
    </div>
  );
}
