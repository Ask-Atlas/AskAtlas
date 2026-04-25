"use client";

import { History, Loader2, Send, X } from "lucide-react";
import {
  type FormEvent,
  type KeyboardEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

import {
  addRecentPrompt,
  getRecentPrompts,
  RECENT_PROMPTS_LIMIT,
} from "./recent-prompts";
import type {
  AiEditStreamError,
  AiEditStreamStatus,
} from "./use-ai-edit-stream";

interface AskAiPopoverProps {
  status: AiEditStreamStatus;
  replacement: string;
  error: AiEditStreamError | null;
  /** Submit triggers the SSE call. Caller owns selection capture. */
  onSubmit: (instruction: string) => void;
  /** Cancels in-flight request and closes the popover. */
  onCancel: () => void;
}

/**
 * Compact prompt input rendered inside the editor's bubble menu when
 * the user clicks "Ask AI". The popover container (Radix Popover)
 * lives in the bubble-menu component; this is just the body.
 *
 * Lifecycle (kept here so the bubble menu's render stays clean):
 *   - On mount, focus the input + load recent prompts.
 *   - Esc cancels (caller closes the popover).
 *   - Submit fires `onSubmit(instruction)` and persists the prompt
 *     to localStorage so the recents dropdown reflects it next open.
 *   - While `status === "streaming"`, show a loading affordance and
 *     a live-streamed preview of the replacement so the user sees
 *     the model is working. ASK-217 will replace this preview block
 *     with the diff overlay; this component intentionally does NOT
 *     write the replacement back into the editor.
 */
export function AskAiPopover({
  status,
  replacement,
  error,
  onSubmit,
  onCancel,
}: AskAiPopoverProps) {
  const [instruction, setInstruction] = useState("");
  // Lazy initializer reads localStorage once on mount. SSR-safe via
  // the helper's `typeof window` guard.
  const [recents, setRecents] = useState<string[]>(() => getRecentPrompts());
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Defer focus to the next frame so Radix's autoFocus -> input
    // hand-off doesn't race with the popover's open animation.
    const id = requestAnimationFrame(() => inputRef.current?.focus());
    return () => cancelAnimationFrame(id);
  }, []);

  const isStreaming = status === "streaming";
  const isDone = status === "done";
  const isError = status === "error";
  const hasError = isError && error !== null;

  const trimmed = useMemo(() => instruction.trim(), [instruction]);
  const canSubmit = trimmed.length > 0 && !isStreaming;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!canSubmit) return;
    setRecents(addRecentPrompt(trimmed));
    onSubmit(trimmed);
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      onCancel();
    }
  };

  return (
    <div className="flex w-80 flex-col gap-3" role="group" aria-label="Ask AI">
      <form onSubmit={handleSubmit} className="flex flex-col gap-2">
        <div className="flex items-center gap-2">
          <Input
            ref={inputRef}
            value={instruction}
            onChange={(e) => setInstruction(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask AI to rewrite the selection…"
            aria-label="AI edit instruction"
            disabled={isStreaming}
            autoComplete="off"
            spellCheck
          />
          {recents.length > 0 && !isStreaming ? (
            <RecentPromptsMenu
              prompts={recents}
              onPick={(p) => setInstruction(p)}
            />
          ) : null}
        </div>
        <div className="flex items-center justify-between gap-2">
          <p className="text-muted-foreground text-xs">
            {isStreaming
              ? "Streaming…"
              : isDone
                ? "Done · diff overlay coming next"
                : "Press ⏎ to send · Esc to close"}
          </p>
          <div className="flex items-center gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onCancel}
              aria-label={isStreaming ? "Stop AI edit" : "Close Ask AI"}
            >
              {isStreaming ? "Stop" : <X className="size-4" />}
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={!canSubmit}
              aria-label="Submit AI edit"
            >
              {isStreaming ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Send className="size-4" />
              )}
            </Button>
          </div>
        </div>
      </form>

      {(isStreaming || isDone) && replacement.length > 0 ? (
        <PreviewPanel replacement={replacement} streaming={isStreaming} />
      ) : null}

      {hasError ? (
        <p
          role="alert"
          className="bg-destructive/10 text-destructive rounded-md px-2 py-1.5 text-xs"
        >
          {error.message}
        </p>
      ) : null}
    </div>
  );
}

interface PreviewPanelProps {
  replacement: string;
  streaming: boolean;
}

function PreviewPanel({ replacement, streaming }: PreviewPanelProps) {
  return (
    <div
      className={cn(
        "border-border bg-muted/40 rounded-md border px-3 py-2",
        streaming && "ai-streaming-pulse",
      )}
      aria-live="polite"
      aria-busy={streaming}
    >
      <p className="text-muted-foreground mb-1 text-[11px] font-medium tracking-wide uppercase">
        Preview
      </p>
      <p className="max-h-40 overflow-y-auto text-sm leading-relaxed whitespace-pre-wrap">
        {replacement}
        {streaming ? (
          <span
            aria-hidden
            className="bg-foreground/70 ml-0.5 inline-block h-3 w-1 animate-pulse align-baseline"
          />
        ) : null}
      </p>
    </div>
  );
}

interface RecentPromptsMenuProps {
  prompts: string[];
  onPick: (prompt: string) => void;
}

function RecentPromptsMenu({ prompts, onPick }: RecentPromptsMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`Recent prompts (last ${RECENT_PROMPTS_LIMIT})`}
        >
          <History className="size-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="max-w-72">
        {prompts.map((prompt) => (
          <DropdownMenuItem
            key={prompt}
            onSelect={() => onPick(prompt)}
            className="line-clamp-2 text-sm"
          >
            {prompt}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
