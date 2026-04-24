"use client";

import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { useEffect, useMemo } from "react";
import { Markdown } from "tiptap-markdown";

import type { RefSummary } from "@/lib/api/types";
import { cn } from "@/lib/utils";

import { EntityRefProvider } from "../refs/entity-ref-context";
import { extractRefs } from "../refs/extract-refs";
import { rewritePastedUrl } from "../paste-rewriter";

import { EntityRefNode, type EntityType } from "./entity-ref-node";
import { preprocessMarkdown } from "./markdown-codec";

interface TiptapEditorProps {
  value: string;
  onChange: (md: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  allowedHosts?: string[];
  initialRefs?: Record<string, RefSummary | null>;
}

function resolveHosts(explicit?: string[]): string[] {
  if (explicit && explicit.length > 0) return explicit;
  const hosts: string[] = [];
  const env = process.env.NEXT_PUBLIC_APP_HOST;
  if (env) hosts.push(env);
  if (typeof window !== "undefined" && window.location?.hostname) {
    hosts.push(window.location.hostname);
  }
  return hosts;
}


export function TiptapEditor({
  value,
  onChange,
  placeholder = "Write in markdown… paste an app URL to embed a live card.",
  disabled,
  className,
  allowedHosts,
  initialRefs,
}: TiptapEditorProps) {
  const hosts = useMemo(() => resolveHosts(allowedHosts), [allowedHosts]);
  const refs = useMemo(() => extractRefs(value), [value]);

  const editor = useEditor({
    immediatelyRender: false,
    editable: !disabled,
    extensions: [
      StarterKit,
      Link.configure({ openOnClick: false, autolink: true }),
      Placeholder.configure({ placeholder }),
      Markdown.configure({
        html: true,
        tightLists: true,
        transformPastedText: false,
        transformCopiedText: false,
      }),
      EntityRefNode,
    ],
    content: preprocessMarkdown(value),
    editorProps: {
      attributes: {
        class: cn(
          "prose prose-neutral dark:prose-invert max-w-none min-h-56 rounded-md border p-3 focus:outline-none",
          disabled && "cursor-not-allowed opacity-70",
        ),
      },
      handlePaste(view, event) {
        const text = event.clipboardData?.getData("text/plain") ?? "";
        const rewrite = rewritePastedUrl(text, hosts);
        if (!rewrite) return false;
        const { state } = view;
        const inlineCtx =
          state.selection.$from.parent.type.name === "paragraph" &&
          state.selection.$from.parent.content.size > 0;
        const variant: "inline" | "leaf" = inlineCtx ? "inline" : "leaf";
        event.preventDefault();
        const node = state.schema.nodes.entityRef?.create({
          type: rewrite.type as EntityType,
          id: rewrite.id,
          variant,
        });
        if (!node) return false;
        const tr = state.tr.replaceSelectionWith(node, false);
        view.dispatch(tr);
        return true;
      },
    },
  });

  useEffect(() => {
    if (!editor) return;
    const current =
      (editor.storage as { markdown?: { getMarkdown?: () => string } })
        .markdown?.getMarkdown?.() ?? "";
    if (current === value) return;
    editor.commands.setContent(preprocessMarkdown(value), {
      emitUpdate: false,
    });
  }, [editor, value]);

  useEffect(() => {
    if (!editor) return;
    const handler = () => {
      const md =
        (editor.storage as { markdown?: { getMarkdown?: () => string } })
          .markdown?.getMarkdown?.() ?? "";
      onChange(md);
    };
    editor.on("update", handler);
    return () => {
      editor.off("update", handler);
    };
  }, [editor, onChange]);

  return (
    <EntityRefProvider refs={refs} initial={initialRefs}>
      <EditorContent editor={editor} className={className} />
    </EntityRefProvider>
  );
}
