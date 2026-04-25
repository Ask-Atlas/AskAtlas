"use client";

import { Loader2 } from "lucide-react";
import { useOptimistic, useState, useTransition } from "react";

import { Button } from "@/components/ui/button";
import { ConfirmationDialog } from "@/lib/features/shared/confirmation-dialog";
import { cn } from "@/lib/utils";

type MembershipState = "member" | "not-member" | "unknown";

interface SectionMembershipButtonProps {
  membership: MembershipState;
  /**
   * Fires the actual join request. Rejections are caught internally so
   * the optimistic state reverts on settle; callers surface errors
   * through their own toast (so different surfaces -- detail page vs
   * onboarding -- can phrase the failure in context).
   */
  onJoin: () => Promise<void>;
  /** Fires after the leave confirmation dialog is confirmed. Same error contract as `onJoin`. */
  onLeave: () => Promise<void>;
  className?: string;
}

export function SectionMembershipButton({
  membership,
  onJoin,
  onLeave,
  className,
}: SectionMembershipButtonProps) {
  const [optimisticMembership, setOptimisticMembership] = useOptimistic(
    membership,
    (_current, next: MembershipState) => next,
  );
  const [isPending, startTransition] = useTransition();
  const [confirmOpen, setConfirmOpen] = useState(false);

  if (optimisticMembership === "unknown") {
    return (
      <Button
        type="button"
        variant="outline"
        disabled
        aria-label="Checking enrollment"
        className={className}
      >
        <Loader2 className="size-4 animate-spin" aria-hidden />
      </Button>
    );
  }

  if (optimisticMembership === "member") {
    const handleConfirmLeave = () => {
      startTransition(async () => {
        setOptimisticMembership("not-member");
        try {
          await onLeave();
        } catch {
          // useOptimistic reverts on settle; caller surfaces toast.
        } finally {
          setConfirmOpen(false);
        }
      });
    };

    return (
      <>
        <Button
          type="button"
          variant="outline"
          disabled={isPending}
          onClick={() => setConfirmOpen(true)}
          className={cn("min-w-[5.5rem]", className)}
        >
          Enrolled
        </Button>
        <ConfirmationDialog
          open={confirmOpen}
          onOpenChange={setConfirmOpen}
          title="Leave this section?"
          description="You'll lose access to section members and course announcements. You can rejoin later."
          confirmLabel={isPending ? "Leaving…" : "Leave"}
          cancelLabel="Cancel"
          destructive
          disabled={isPending}
          onConfirm={handleConfirmLeave}
        />
      </>
    );
  }

  const handleJoin = () => {
    // Don't optimistically swap to "member" -- access to the course
    // material depends on the server confirming enrollment, so the
    // caller drives a router.refresh inside this transition and the
    // button stays in `isPending` (spinner) until the new server
    // payload renders. Errors bubble to the caller's toast.
    startTransition(async () => {
      try {
        await onJoin();
      } catch {
        // caller surfaces toast.
      }
    });
  };

  return (
    <Button
      type="button"
      disabled={isPending}
      onClick={handleJoin}
      aria-label={isPending ? "Joining section" : undefined}
      className={cn("min-w-[5.5rem]", className)}
    >
      {isPending ? (
        <Loader2 className="size-4 animate-spin" aria-hidden={true} />
      ) : (
        "Join"
      )}
    </Button>
  );
}
