"use client";

import { useRouter } from "next/navigation";
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
  const router = useRouter();
  const [membership, setMembership] = useState<Membership>(initialMembership);

  const sectionLabel = section.section_code
    ? `Section ${section.section_code}`
    : "Section";

  const handleJoin = async () => {
    try {
      await joinSection(courseId, section.id);
      // Re-fetch the route so the page swaps from the picker view to
      // the enrolled banner + study-guide grid. router.refresh() is
      // called inside the button's transition, so the spinner stays
      // up until the new server payload renders -- no flash of
      // stale "Join" UI between success and the page swap.
      router.refresh();
    } catch (err) {
      toast.error(err);
      throw err;
    }
  };

  const handleLeave = async () => {
    try {
      await leaveSection(courseId, section.id);
      setMembership("not-member");
      router.refresh();
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
        <span className="bg-muted text-foreground rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold tracking-[-0.2px]">
          {sectionLabel}
        </span>
        <div className="flex min-w-0 flex-col gap-0.5">
          <p className="text-foreground truncate text-[14px] font-medium">
            {section.instructor_name ?? "Instructor TBA"}
          </p>
          <p className="text-muted-foreground truncate text-[12px]">
            <span>{section.term}</span>
            <span className="text-muted-foreground/50 px-2" aria-hidden={true}>
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
