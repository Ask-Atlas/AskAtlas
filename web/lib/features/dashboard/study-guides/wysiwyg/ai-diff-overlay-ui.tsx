"use client";

import type { Editor } from "@tiptap/core";
import { Check, CheckCheck, Loader2, X, XCircle } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";

import {
  aiDiffPluginKey,
  type AiDiffPluginHunk,
  type AiDiffPluginState,
} from "./ai-diff-overlay";

interface AiDiffOverlayUiProps {
  editor: Editor;
  /**
   * Called when the user resolves every hunk -- either via per-hunk
   * decisions, "Accept all" / "Reject all", or Esc-cancel. The
   * `accepted` flag is the overall outcome for the audit row PATCH:
   * true if the user kept any hunk's AI text, false otherwise.
   */
  onResolved: (accepted: boolean) => void;
}

/**
 * Diff review UI rendered inside the Ask AI popover (ASK-217). Reads
 * hunk state from the AiDiffOverlay ProseMirror plugin and dispatches
 * accept/reject commands back into the editor.
 *
 * Resolution policy:
 *   - per-hunk accept -> tracks "anyAccepted = true" so the final
 *     PATCH records the user took at least one suggestion
 *   - per-hunk reject -> no flip
 *   - Accept all -> anyAccepted = true (assuming there were hunks)
 *   - Reject all -> anyAccepted stays where it was; if it was false
 *     the resolution PATCH records `accepted: false`
 */
export function AiDiffOverlayUi({ editor, onResolved }: AiDiffOverlayUiProps) {
  const [pluginState, setPluginState] = useState<AiDiffPluginState | null>(
    () => aiDiffPluginKey.getState(editor.state) ?? null,
  );
  const anyAcceptedRef = useRef(false);
  const resolvedFiredRef = useRef(false);

  // Mirror the plugin state into React on every editor transaction
  // (the plugin is the source of truth -- this is a one-way mirror).
  useEffect(() => {
    const sync = () => {
      setPluginState(aiDiffPluginKey.getState(editor.state) ?? null);
    };
    editor.on("transaction", sync);
    return () => {
      editor.off("transaction", sync);
    };
  }, [editor]);

  // Fire `onResolved` exactly once per review session, when the last
  // hunk is gone. resolvedFiredRef guards against the empty-list
  // mirror running twice (e.g. acceptAll dispatches several
  // transactions in sequence).
  useEffect(() => {
    if (resolvedFiredRef.current) return;
    if (!pluginState) return;
    if (pluginState.hunks.length > 0) return;
    resolvedFiredRef.current = true;
    onResolved(anyAcceptedRef.current);
  }, [pluginState, onResolved]);

  const handleAccept = useCallback(
    (id: string) => {
      anyAcceptedRef.current = true;
      editor.chain().focus().acceptAiHunk(id).run();
    },
    [editor],
  );

  const handleReject = useCallback(
    (id: string) => {
      editor.chain().focus().rejectAiHunk(id).run();
    },
    [editor],
  );

  const handleAcceptAll = useCallback(() => {
    if (!pluginState || pluginState.hunks.length === 0) return;
    anyAcceptedRef.current = true;
    editor.chain().focus().acceptAllAiHunks().run();
  }, [editor, pluginState]);

  const handleRejectAll = useCallback(() => {
    if (!pluginState || pluginState.hunks.length === 0) return;
    editor.chain().focus().rejectAllAiHunks().run();
  }, [editor, pluginState]);

  if (!pluginState) {
    return null;
  }

  const total = pluginState.hunks.length;
  if (total === 0) {
    return (
      <div
        className="text-muted-foreground flex items-center gap-2 text-xs"
        aria-live="polite"
      >
        <Loader2 className="size-3 animate-spin" />
        Resolving…
      </div>
    );
  }

  return (
    <div
      className="flex w-80 flex-col gap-3"
      role="group"
      aria-label="Review AI edit"
    >
      <div className="flex items-center justify-between gap-2">
        <p className="text-muted-foreground text-xs font-medium">
          {total} {total === 1 ? "change" : "changes"} to review
        </p>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleRejectAll}
            aria-label="Reject all changes"
            className="gap-1.5"
          >
            <XCircle className="size-3.5" />
            <span className="text-xs">Reject all</span>
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={handleAcceptAll}
            aria-label="Accept all changes"
            className="gap-1.5"
          >
            <CheckCheck className="size-3.5" />
            <span className="text-xs">Accept all</span>
          </Button>
        </div>
      </div>

      <ul
        className="border-border bg-muted/40 flex max-h-56 flex-col gap-1 overflow-y-auto rounded-md border p-1"
        aria-label="Pending changes"
      >
        {pluginState.hunks.map((hunk) => (
          <HunkRow
            key={hunk.id}
            hunk={hunk}
            onAccept={handleAccept}
            onReject={handleReject}
          />
        ))}
      </ul>
    </div>
  );
}

interface HunkRowProps {
  hunk: AiDiffPluginHunk;
  onAccept: (id: string) => void;
  onReject: (id: string) => void;
}

function HunkRow({ hunk, onAccept, onReject }: HunkRowProps) {
  const label =
    hunk.op === "insert"
      ? "Insertion"
      : hunk.op === "delete"
        ? "Deletion"
        : "Replacement";
  const ariaSummary = hunkSummary(hunk);
  return (
    <li className="bg-background flex items-start gap-2 rounded px-2 py-1.5 text-xs">
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="text-muted-foreground text-[10px] font-medium tracking-wide uppercase">
          {label}
        </span>
        <HunkTextPreview hunk={hunk} />
      </div>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="text-destructive size-6 shrink-0"
        onClick={() => onReject(hunk.id)}
        aria-label={`Reject ${ariaSummary}`}
      >
        <X className="size-3.5" />
      </Button>
      <Button
        type="button"
        size="icon"
        className="size-6 shrink-0"
        onClick={() => onAccept(hunk.id)}
        aria-label={`Accept ${ariaSummary}`}
      >
        <Check className="size-3.5" />
      </Button>
    </li>
  );
}

function HunkTextPreview({ hunk }: { hunk: AiDiffPluginHunk }) {
  // Long-form rewrites can be a whole paragraph. Cap at 240 chars so
  // memory stays bounded, then let CSS handle wrap + internal scroll
  // so the row never spills past the popover. The inline diff in the
  // editor still shows the full text untruncated.
  const del = trimSnippet(hunk.delText);
  const ins = trimSnippet(hunk.insText);
  // `min-w-0` + `break-words` is the combo that actually lets long
  // unbroken strings wrap inside a flex child. `max-h-32 overflow-y-auto`
  // is a per-row safety net for genuinely huge hunks.
  const previewClass =
    "font-mono text-xs leading-snug min-w-0 break-words whitespace-pre-wrap max-h-32 overflow-y-auto";
  if (hunk.op === "insert") {
    return (
      <span className={previewClass}>
        <span className="text-emerald-700 dark:text-emerald-400">+ {ins}</span>
      </span>
    );
  }
  if (hunk.op === "delete") {
    return (
      <span className={previewClass}>
        <span className="text-rose-700 line-through dark:text-rose-400">
          − {del}
        </span>
      </span>
    );
  }
  return (
    <span className={previewClass}>
      <span className="text-rose-700 line-through dark:text-rose-400">
        {del}
      </span>{" "}
      <span aria-hidden className="text-muted-foreground">
        →
      </span>{" "}
      <span className="text-emerald-700 dark:text-emerald-400">{ins}</span>
    </span>
  );
}

function hunkSummary(hunk: AiDiffPluginHunk): string {
  const del = trimSnippet(hunk.delText, 40);
  const ins = trimSnippet(hunk.insText, 40);
  if (hunk.op === "insert") return `insert "${ins}"`;
  if (hunk.op === "delete") return `delete "${del}"`;
  return `replace "${del}" with "${ins}"`;
}

function trimSnippet(text: string, max = 240): string {
  if (text.length <= max) return text || "(empty)";
  return `${text.slice(0, max - 1)}…`;
}
