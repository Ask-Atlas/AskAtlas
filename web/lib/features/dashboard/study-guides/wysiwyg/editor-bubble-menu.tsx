"use client";

import { useAuth } from "@clerk/nextjs";
import { posToDOMRect } from "@tiptap/core";
import type { Editor } from "@tiptap/core";
import { BubbleMenu } from "@tiptap/react/menus";
import { Bold, Code, Italic, Sparkles, Strikethrough } from "lucide-react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

import { API_BASE } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverAnchor,
  PopoverContent,
} from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

import { findPreset } from "../ai/presets";
import { AiDiffOverlayUi } from "./ai-diff-overlay-ui";
import { aiDiffPluginKey } from "./ai-diff-overlay";
import {
  aiSelectionPluginKey,
  type AiSelectionStatus,
} from "./ai-selection-range";
import { AskAiPopover } from "./ask-ai-popover";
import { useAiEditStream } from "./use-ai-edit-stream";

export interface EditorBubbleMenuProps {
  editor: Editor;
  /**
   * Provide to enable the "Ask AI" entry. Omitted in create-mode (no
   * guide id yet) -- the menu still renders formatting buttons.
   */
  aiEdit?: {
    guideId: string;
    title?: string;
  };
}

const CONTEXT_WINDOW = 500;

/**
 * Floating selection toolbar (TipTap v3 BubbleMenu, Floating UI under
 * the hood). Two-phase popover lifecycle:
 *
 *   1. Prompt phase (ASK-216): user submits an instruction; the SSE
 *      stream returns a `replacement` and the persisted audit row's
 *      `editId`.
 *   2. Diff review phase (ASK-217): on stream `done`, we seed the
 *      AiDiffOverlay ProseMirror plugin which replaces the selection
 *      range with a merged ins/del fragment in the doc. The popover
 *      swaps in `<AiDiffOverlayUi>` for per-hunk accept/reject. On
 *      resolution we PATCH the audit row with the overall outcome.
 *
 * The visible "you're editing this text" highlight is owned by the
 * AiSelectionRange ProseMirror plugin during the prompt phase; once
 * the diff is seeded, the marks themselves provide the visual.
 */
export function EditorBubbleMenu({ editor, aiEdit }: EditorBubbleMenuProps) {
  const [askOpen, setAskOpen] = useState(false);
  const [target, setTarget] = useState<{
    from: number;
    to: number;
    text: string;
    spansMultipleBlocks: boolean;
  } | null>(null);
  const [anchorRect, setAnchorRect] = useState<DOMRect | null>(null);
  const [inDiffReview, setInDiffReview] = useState(false);
  // Capture the editId once we transition into diff review so we can
  // PATCH it on resolution even after the stream hook resets.
  const editIdRef = useRef<string | null>(null);
  // Avoid double-firing the seed effect if React re-runs the effect
  // for a transient reason after a successful seed.
  const diffSeededRef = useRef(false);
  // Tracks the preset id of the chip that kicked off this stream
  // (null for free-form prompts). Read in the seed effect so the
  // preset's transformReplacement (e.g. TL;DR's prepend) can run
  // before the diff overlay is seeded.
  const activePresetIdRef = useRef<string | null>(null);

  const stream = useAiEditStream({ guideId: aiEdit?.guideId ?? "" });
  const { status, replacement, error, editId, start, cancel, reset } = stream;

  // Browser-side PATCH for the diff resolution. Avoids the server-
  // action chain (lib/api/server-client.ts -> @clerk/nextjs/server)
  // so this component stays bundle-clean for storybook + tests.
  const { getToken } = useAuth();
  const patchAiEdit = useCallback(
    async (guideId: string, id: string, accepted: boolean) => {
      let token: string | null = null;
      try {
        token = await getToken();
      } catch {
        token = null;
      }
      await fetch(`${API_BASE}/study-guides/${guideId}/ai/edits/${id}`, {
        method: "PATCH",
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ accepted }),
      });
    },
    [getToken],
  );

  const shouldShow = useCallback(
    ({ state }: { state: { selection: { empty: boolean } } }) => {
      if (askOpen) return true;
      return !state.selection.empty;
    },
    [askOpen],
  );

  const setHighlight = useCallback(
    (next: { from: number; to: number; status: AiSelectionStatus } | null) => {
      const { state, dispatch } = editor.view;
      dispatch(state.tr.setMeta(aiSelectionPluginKey, next));
    },
    [editor],
  );

  const handleOpenAsk = () => {
    if (!aiEdit) return;
    const captured = captureSelection(editor);
    if (!captured) return;
    diffSeededRef.current = false;
    editIdRef.current = null;
    activePresetIdRef.current = null;
    setTarget({
      from: captured.from,
      to: captured.to,
      text: captured.text,
      spansMultipleBlocks: captured.spansMultipleBlocks,
    });
    setAnchorRect(captured.rect);
    setHighlight({ from: captured.from, to: captured.to, status: "idle" });
    setInDiffReview(false);
    reset();
    setAskOpen(true);
  };

  const closeAsk = useCallback(() => {
    cancel();
    // If a diff was seeded but never resolved, drop any del-marked
    // text + clear the plugin state so a half-review doesn't bleed
    // into the next save.
    if (aiDiffPluginKey.getState(editor.state)) {
      editor.chain().focus().clearAiDiff().run();
    }
    diffSeededRef.current = false;
    editIdRef.current = null;
    activePresetIdRef.current = null;
    setInDiffReview(false);
    setAskOpen(false);
    setTarget(null);
    setAnchorRect(null);
    setHighlight(null);
    reset();
  }, [cancel, reset, setHighlight, editor]);

  const handleSubmit = (instruction: string, presetId?: string) => {
    if (!aiEdit || !target) return;
    const live = aiSelectionPluginKey.getState(editor.state);
    const from = live?.from ?? target.from;
    const to = live?.to ?? target.to;
    const text =
      live && live.from !== target.from
        ? editor.state.doc.textBetween(from, to, "\n", "\n")
        : target.text;
    const docContext = buildDocContext(editor, { from, to }, aiEdit.title);
    activePresetIdRef.current = presetId ?? null;
    void start({
      selectionText: text,
      selectionStart: from,
      selectionEnd: to,
      instruction,
      docContext,
    });
  };

  // Mirror stream status into the highlight so the decoration pulses
  // while the model streams and stops once we hit done / error.
  useEffect(() => {
    if (!askOpen) return;
    if (inDiffReview) return; // highlight handed off to the diff marks
    const live = aiSelectionPluginKey.getState(editor.state);
    if (!live) return;
    setHighlight({
      from: live.from,
      to: live.to,
      status: status === "streaming" ? "streaming" : "idle",
    });
  }, [askOpen, status, setHighlight, editor, inDiffReview]);

  // On stream `done`, seed the diff overlay. We pull live (mapped)
  // positions from the selection-range plugin so concurrent edits
  // upstream of the selection don't put the diff in the wrong place.
  /* eslint-disable react-hooks/set-state-in-effect --
   * The setInDiffReview / setHighlight calls below are the response
   * to an external state transition (stream status -> "done") and are
   * guarded by diffSeededRef so they can't cascade.
   */
  useEffect(() => {
    if (status !== "done") return;
    if (!target) return;
    if (replacement.length === 0) return;
    if (diffSeededRef.current) return;
    diffSeededRef.current = true;
    const live = aiSelectionPluginKey.getState(editor.state);
    const from = live?.from ?? target.from;
    const to = live?.to ?? target.to;
    const originalText =
      live && live.from !== target.from
        ? editor.state.doc.textBetween(from, to, "\n", "\n")
        : target.text;
    editIdRef.current = editId;
    // Apply any preset-specific client-side transform BEFORE seeding
    // the diff. TL;DR uses this to wrap the model's reply as a
    // blockquote prepended to the original instead of replacing it.
    const presetId = activePresetIdRef.current;
    const preset = presetId ? findPreset(presetId) : undefined;
    const finalReplacement =
      preset?.transformReplacement?.(replacement, originalText) ?? replacement;
    const ok = editor
      .chain()
      .focus()
      .seedAiDiff({
        editId,
        originalText,
        replacement: finalReplacement,
        from,
        to,
      })
      .run();
    if (ok) {
      setHighlight(null);
      setInDiffReview(true);
    } else {
      closeAsk();
    }
  }, [status, target, replacement, editId, editor, closeAsk, setHighlight]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleResolved = useCallback(
    (accepted: boolean) => {
      const guideId = aiEdit?.guideId;
      const id = editIdRef.current;
      // Best-effort PATCH: a failure here loses the eval signal for
      // this edit, but the user's accept/reject is already reflected
      // in the doc, so we don't show a blocking error.
      if (guideId && id) {
        void patchAiEdit(guideId, id, accepted).catch(() => {});
      }
      closeAsk();
    },
    [aiEdit?.guideId, closeAsk, patchAiEdit],
  );

  // Recompute the popover anchor rect on scroll, resize, AND every
  // editor transaction so the popover stays glued to the highlighted
  // range even after typing into the doc.
  useLayoutEffect(() => {
    if (!askOpen || !target) return;
    const update = () => {
      const live = aiSelectionPluginKey.getState(editor.state);
      const from = live?.from ?? target.from;
      const to = live?.to ?? target.to;
      try {
        setAnchorRect(posToDOMRect(editor.view, from, to));
      } catch {
        // Doc was edited out from under us; ignore.
      }
    };
    update();
    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);
    editor.on("transaction", update);
    return () => {
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
      editor.off("transaction", update);
    };
  }, [askOpen, target, editor]);

  // Cancel any in-flight stream + clear highlight when unmounting.
  useEffect(() => {
    return () => {
      cancel();
      setHighlight(null);
    };
  }, [cancel, setHighlight]);

  return (
    <>
      <BubbleMenu
        editor={editor}
        shouldShow={shouldShow}
        options={{ placement: "top", offset: 8 }}
        className="bg-popover text-popover-foreground flex items-center gap-0.5 rounded-md border p-1 shadow-md"
      >
        <FormatButton
          editor={editor}
          mark="bold"
          aria-label="Bold"
          icon={<Bold className="size-4" />}
        />
        <FormatButton
          editor={editor}
          mark="italic"
          aria-label="Italic"
          icon={<Italic className="size-4" />}
        />
        <FormatButton
          editor={editor}
          mark="strike"
          aria-label="Strikethrough"
          icon={<Strikethrough className="size-4" />}
        />
        <FormatButton
          editor={editor}
          mark="code"
          aria-label="Inline code"
          icon={<Code className="size-4" />}
        />
        {aiEdit ? (
          <>
            <Separator orientation="vertical" className="mx-1 h-5" />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onMouseDown={(e) => e.preventDefault()}
              onClick={handleOpenAsk}
              className="gap-1.5"
              aria-label="Ask AI to rewrite the selection"
            >
              <Sparkles className="size-4" />
              <span className="text-xs font-medium">Ask AI</span>
            </Button>
          </>
        ) : null}
      </BubbleMenu>

      {aiEdit ? (
        <Popover
          open={askOpen}
          onOpenChange={(next) => {
            if (!next) closeAsk();
          }}
        >
          {anchorRect ? (
            <PopoverAnchor asChild>
              <SelectionAnchor rect={anchorRect} />
            </PopoverAnchor>
          ) : null}
          <PopoverContent
            side="bottom"
            align="start"
            sideOffset={8}
            className="w-auto p-3"
            onEscapeKeyDown={(e) => {
              e.preventDefault();
              closeAsk();
            }}
            onInteractOutside={(e) => {
              const target = e.detail.originalEvent.target;
              if (
                target instanceof Element &&
                target.closest('[role="menu"], [role="listbox"]')
              ) {
                e.preventDefault();
                return;
              }
              if (e.detail.originalEvent.type === "pointerdown") {
                closeAsk();
              }
            }}
          >
            {inDiffReview ? (
              <AiDiffOverlayUi editor={editor} onResolved={handleResolved} />
            ) : (
              <AskAiPopover
                status={status}
                replacement={replacement}
                error={error}
                onSubmit={handleSubmit}
                onCancel={closeAsk}
                selectionSpansMultipleBlocks={
                  target?.spansMultipleBlocks ?? false
                }
              />
            )}
          </PopoverContent>
        </Popover>
      ) : null}
    </>
  );
}

interface FormatButtonProps {
  editor: Editor;
  mark: "bold" | "italic" | "strike" | "code";
  icon: React.ReactNode;
  "aria-label": string;
}

function FormatButton({
  editor,
  mark,
  icon,
  "aria-label": label,
}: FormatButtonProps) {
  const isActive = editor.isActive(mark);
  const handleClick = () => {
    const chain = editor.chain().focus();
    switch (mark) {
      case "bold":
        chain.toggleBold().run();
        break;
      case "italic":
        chain.toggleItalic().run();
        break;
      case "strike":
        chain.toggleStrike().run();
        break;
      case "code":
        chain.toggleCode().run();
        break;
    }
  };
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onMouseDown={(e) => e.preventDefault()}
      onClick={handleClick}
      aria-pressed={isActive}
      aria-label={label}
      className={cn("size-8", isActive && "bg-accent text-accent-foreground")}
    >
      {icon}
    </Button>
  );
}

interface SelectionAnchorProps {
  rect: DOMRect;
}

const SelectionAnchor = forwardRef<HTMLDivElement, SelectionAnchorProps>(
  function SelectionAnchor({ rect }, ref) {
    return (
      <div
        ref={ref}
        aria-hidden
        style={{
          position: "fixed",
          top: rect.top,
          left: rect.left,
          width: Math.max(rect.width, 1),
          height: Math.max(rect.height, 1),
          pointerEvents: "none",
        }}
      />
    );
  },
);

interface CapturedSelection {
  from: number;
  to: number;
  text: string;
  rect: DOMRect;
  /**
   * Whether the selection straddles a block boundary (paragraph,
   * heading, list item, etc.). Drives multi-paragraph-only preset
   * eligibility (ASK-218 "Reorganize"). `textBetween` emits "\n" at
   * each block boundary so a newline in the captured text means we
   * spanned more than one block.
   */
  spansMultipleBlocks: boolean;
}

function captureSelection(editor: Editor): CapturedSelection | null {
  const { state } = editor;
  const { from, to } = state.selection;
  if (from === to) return null;
  const text = state.doc.textBetween(from, to, "\n", "\n");
  if (!text.trim()) return null;
  let rect: DOMRect;
  try {
    rect = posToDOMRect(editor.view, from, to);
  } catch {
    return null;
  }
  return { from, to, text, rect, spansMultipleBlocks: text.includes("\n") };
}

function buildDocContext(
  editor: Editor,
  target: { from: number; to: number },
  title?: string,
): { title?: string; preceding?: string; following?: string } | undefined {
  const { doc } = editor.state;
  const docSize = doc.content.size;
  const precedingFrom = Math.max(0, target.from - CONTEXT_WINDOW);
  const followingTo = Math.min(docSize, target.to + CONTEXT_WINDOW);
  const preceding = doc.textBetween(precedingFrom, target.from, "\n", "\n");
  const following = doc.textBetween(target.to, followingTo, "\n", "\n");
  const ctx: { title?: string; preceding?: string; following?: string } = {};
  if (title && title.trim()) ctx.title = title.trim();
  if (preceding) ctx.preceding = preceding;
  if (following) ctx.following = following;
  return Object.keys(ctx).length > 0 ? ctx : undefined;
}
