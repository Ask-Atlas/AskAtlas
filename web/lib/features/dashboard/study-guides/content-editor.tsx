"use client";

import {
  forwardRef,
  useCallback,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from "react";
import type { ClipboardEvent, ChangeEvent, TextareaHTMLAttributes } from "react";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

import { ArticleRenderer } from "./article-renderer";
import { rewritePastedUrl } from "./paste-rewriter";

interface ContentEditorProps
  extends Omit<
    TextareaHTMLAttributes<HTMLTextAreaElement>,
    "value" | "onChange"
  > {
  value: string;
  onChange: (next: string) => void;
  allowedHosts?: string[];
}

function resolveAllowedHosts(explicit?: string[]): string[] {
  if (explicit && explicit.length > 0) return explicit;
  const envHost = process.env.NEXT_PUBLIC_APP_HOST;
  const hosts: string[] = [];
  if (envHost) hosts.push(envHost);
  if (typeof window !== "undefined" && window.location?.hostname) {
    hosts.push(window.location.hostname);
  }
  return hosts;
}

export const ContentEditor = forwardRef<
  HTMLTextAreaElement,
  ContentEditorProps
>(function ContentEditor(
  { value, onChange, allowedHosts, className, disabled, ...rest },
  ref,
) {
  const internalRef = useRef<HTMLTextAreaElement>(null);
  useImperativeHandle(ref, () => internalRef.current as HTMLTextAreaElement);

  const [tab, setTab] = useState<"write" | "preview">("write");
  const hosts = useMemo(() => resolveAllowedHosts(allowedHosts), [
    allowedHosts,
  ]);

  const handlePaste = useCallback(
    (e: ClipboardEvent<HTMLTextAreaElement>) => {
      const pasted = e.clipboardData.getData("text/plain");
      const rewrite = rewritePastedUrl(pasted, hosts);
      if (!rewrite) return;
      e.preventDefault();
      const el = internalRef.current;
      if (!el) return;
      // execCommand lands in a single undo step so Ctrl+Z reverts
      // the insertion as one operation.
      const ok =
        typeof document !== "undefined" &&
        document.execCommand?.("insertText", false, rewrite.directive);
      if (!ok) {
        const start = el.selectionStart ?? value.length;
        const end = el.selectionEnd ?? start;
        const next =
          value.slice(0, start) + rewrite.directive + value.slice(end);
        onChange(next);
        requestAnimationFrame(() => {
          el.focus();
          const pos = start + rewrite.directive.length;
          el.setSelectionRange(pos, pos);
        });
      }
    },
    [hosts, onChange, value],
  );

  const handleChange = useCallback(
    (e: ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e.target.value);
    },
    [onChange],
  );

  return (
    <Tabs
      value={tab}
      onValueChange={(v) => setTab(v as "write" | "preview")}
      className={cn("flex flex-col gap-2", className)}
    >
      <TabsList className="self-start">
        <TabsTrigger value="write" disabled={disabled}>
          Write
        </TabsTrigger>
        <TabsTrigger value="preview" disabled={disabled}>
          Preview
        </TabsTrigger>
      </TabsList>
      <TabsContent value="write" className="m-0">
        <Textarea
          {...rest}
          ref={internalRef}
          value={value}
          onChange={handleChange}
          onPaste={handlePaste}
          disabled={disabled}
        />
      </TabsContent>
      <TabsContent value="preview" className="m-0">
        <div
          className={cn(
            "min-h-56 rounded-md border p-4",
            value.trim() === "" &&
              "text-muted-foreground flex items-center justify-center text-sm",
          )}
        >
          {value.trim() === "" ? (
            <span>Nothing to preview yet.</span>
          ) : (
            <ArticleRenderer content={value} />
          )}
        </div>
      </TabsContent>
    </Tabs>
  );
});
