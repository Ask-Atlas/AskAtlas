"use client";

import { useState } from "react";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

import { ArticleRenderer } from "./article-renderer";
import { TiptapEditor } from "./wysiwyg/editor";

interface ContentEditorProps {
  value: string;
  onChange: (next: string) => void;
  onBlur?: () => void;
  name?: string;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  className?: string;
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

export function ContentEditor({
  value,
  onChange,
  onBlur,
  placeholder,
  disabled,
  className,
  allowedHosts,
}: ContentEditorProps) {
  const [tab, setTab] = useState<"write" | "preview">("write");
  const hosts = resolveAllowedHosts(allowedHosts);

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
      <TabsContent value="write" className="m-0" onBlur={onBlur}>
        <TiptapEditor
          value={value}
          onChange={onChange}
          disabled={disabled}
          placeholder={placeholder}
          allowedHosts={hosts}
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
}
