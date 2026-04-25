"use client";

import { TiptapEditor, type AiEditTarget } from "./wysiwyg/editor";

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
  aiEdit?: AiEditTarget;
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
  aiEdit,
}: ContentEditorProps) {
  return (
    <div className={className} onBlur={onBlur}>
      <TiptapEditor
        value={value}
        onChange={onChange}
        disabled={disabled}
        placeholder={placeholder}
        allowedHosts={resolveAllowedHosts(allowedHosts)}
        aiEdit={aiEdit}
      />
    </div>
  );
}
