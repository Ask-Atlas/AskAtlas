"use client";

import { posToDOMRect } from "@tiptap/core";
import type { Editor } from "@tiptap/core";
import { BubbleMenu } from "@tiptap/react/menus";
import { Bold, Code, Italic, Sparkles, Strikethrough } from "lucide-react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useLayoutEffect,
  useState,
} from "react";

import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverAnchor,
  PopoverContent,
} from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

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
 * the hood). Shows on non-empty text selection and exposes:
 *   - Inline formatting toggles (bold / italic / strike / inline code)
 *   - "Ask AI" entry (when {@link aiEdit} is provided) that opens a
 *     Radix Popover anchored to the selection. The popover hosts the
 *     prompt input, recent-prompts dropdown, and streaming preview.
 *
 * The visible "you're editing this text" highlight is owned by the
 * AiSelectionRange ProseMirror plugin (see ./ai-selection-range.ts):
 * we dispatch `setMeta(aiSelectionPluginKey, { from, to, status })`
 * on open and on stream-state transitions, then clear it on close.
 * The plugin renders the highlight as an inline Decoration so it
 * tracks the text on scroll, doc edits, and resize without any JS
 * measurement.
 */
export function EditorBubbleMenu({ editor, aiEdit }: EditorBubbleMenuProps) {
  const [askOpen, setAskOpen] = useState(false);
  // Captured at the moment the user clicks "Ask AI" -- the selection
  // can change once focus moves into the popover input.
  const [target, setTarget] = useState<{
    from: number;
    to: number;
    text: string;
  } | null>(null);
  // Anchor rect for the Radix Popover. Recomputed on scroll / resize
  // via posToDOMRect so the popover follows its target.
  const [anchorRect, setAnchorRect] = useState<DOMRect | null>(null);

  const stream = useAiEditStream({ guideId: aiEdit?.guideId ?? "" });
  const { status, replacement, error, start, cancel, reset } = stream;

  // Keep the bubble menu visible while the popover is open even
  // though the editor has lost focus. Once the popover closes we
  // fall back to the default selection-based visibility rule.
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
    setTarget(captured);
    setAnchorRect(captured.rect);
    setHighlight({ from: captured.from, to: captured.to, status: "idle" });
    reset();
    setAskOpen(true);
  };

  const closeAsk = useCallback(() => {
    cancel();
    setAskOpen(false);
    setTarget(null);
    setAnchorRect(null);
    setHighlight(null);
    reset();
  }, [cancel, reset, setHighlight]);

  const handleSubmit = (instruction: string) => {
    if (!aiEdit || !target) return;
    const docContext = buildDocContext(editor, target, aiEdit.title);
    void start({
      selectionText: target.text,
      selectionStart: target.from,
      selectionEnd: target.to,
      instruction,
      docContext,
    });
  };

  // Mirror stream status into the highlight so the decoration pulses
  // while the model streams and stops once we hit done / error.
  useEffect(() => {
    if (!askOpen || !target) return;
    setHighlight({
      from: target.from,
      to: target.to,
      status: status === "streaming" ? "streaming" : "idle",
    });
  }, [askOpen, target, status, setHighlight]);

  // Recompute the popover anchor rect on scroll / resize so the
  // popover stays glued to the highlighted range. Layout effect so
  // we measure after the DOM commits but before paint.
  useLayoutEffect(() => {
    if (!askOpen || !target) return;
    const update = () => {
      try {
        const rect = posToDOMRect(editor.view, target.from, target.to);
        setAnchorRect(rect);
      } catch {
        // Doc was edited out from under us; ignore.
      }
    };
    update();
    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);
    return () => {
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
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
              // Prevent the editor's blur on mousedown so the doc
              // selection is still intact when our click handler reads
              // it back via captureSelection.
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
              // Only close on click-outside; ignore focus shifts so a
              // recents-dropdown click doesn't dismiss accidentally.
              if (e.detail.originalEvent.type === "pointerdown") {
                closeAsk();
              }
            }}
          >
            <AskAiPopover
              status={status}
              replacement={replacement}
              error={error}
              onSubmit={handleSubmit}
              onCancel={closeAsk}
            />
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
      // Block the default mousedown so the editor's selection isn't
      // collapsed when the user clicks a toolbar button.
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

/**
 * Invisible positioning hint for the Radix Popover. The visible
 * highlight is rendered by the AiSelectionRange ProseMirror plugin
 * as an inline Decoration; this div exists only so Radix has an
 * anchor element it can measure.
 *
 * Ref-forwarding is required because `<PopoverAnchor asChild>` uses
 * Radix's Slot and needs to attach its measurement ref to this node.
 */
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
}

function captureSelection(editor: Editor): CapturedSelection | null {
  const { state } = editor;
  const { from, to } = state.selection;
  if (from === to) return null;
  const text = state.doc.textBetween(from, to, "\n", "\n");
  if (!text.trim()) return null;
  // posToDOMRect is TipTap's blessed helper -- it walks the view's
  // DOM mapping for the [from, to] range so multi-line selections
  // get a sensible bounding rect, unlike coordsAtPos which only
  // returns the cursor box at a single position.
  let rect: DOMRect;
  try {
    rect = posToDOMRect(editor.view, from, to);
  } catch {
    return null;
  }
  return { from, to, text, rect };
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
