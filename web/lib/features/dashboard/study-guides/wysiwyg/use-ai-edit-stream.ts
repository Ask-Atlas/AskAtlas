"use client";

import { useAuth } from "@clerk/nextjs";
import { useCallback, useEffect, useRef, useState } from "react";

import { API_BASE } from "@/lib/api/client";
import type { ApiSchemas } from "@/lib/api/client";

/**
 * Browser-side SSE consumer for `POST /api/study-guides/{id}/ai/edit`.
 *
 * Why a custom hook instead of openapi-fetch:
 *   - openapi-fetch buffers the full body before resolving, which
 *     defeats `text/event-stream` flush-per-chunk semantics.
 *   - Server-side, the route is gated by Clerk; we inject the JWT
 *     here via `useAuth().getToken()` so the same browser fetch can
 *     hit the Go API on a different origin.
 *
 * The hook stays decoupled from the diff overlay (ASK-217). Once the
 * first delta arrives, the consumer (popover today, diff overlay
 * later) drives the UI off `replacement` + `status`.
 */

export type AiEditStreamStatus =
  | "idle"
  | "streaming"
  | "done"
  | "error"
  | "cancelled";

export interface AiEditStreamError {
  message: string;
  status?: number;
}

export interface AiEditStreamUsage {
  inputTokens: number;
  outputTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
}

export interface AiEditStartParams {
  selectionText: string;
  selectionStart: number;
  selectionEnd: number;
  instruction: string;
  docContext?: ApiSchemas["AIEditDocContext"];
}

export interface UseAiEditStreamOptions {
  guideId: string;
}

export interface UseAiEditStreamReturn {
  status: AiEditStreamStatus;
  replacement: string;
  usage: AiEditStreamUsage | null;
  error: AiEditStreamError | null;
  start: (params: AiEditStartParams) => Promise<void>;
  cancel: () => void;
  reset: () => void;
}

interface DeltaPayload {
  text?: string;
}

interface UsagePayload {
  input_tokens?: number;
  output_tokens?: number;
  cache_read_tokens?: number;
  cache_write_tokens?: number;
}

interface ErrorPayload {
  message?: string;
}

const NEWLINE = /\r?\n/;

export function useAiEditStream(
  options: UseAiEditStreamOptions,
): UseAiEditStreamReturn {
  const { guideId } = options;
  const { getToken } = useAuth();

  const [status, setStatus] = useState<AiEditStreamStatus>("idle");
  const [replacement, setReplacement] = useState("");
  const [usage, setUsage] = useState<AiEditStreamUsage | null>(null);
  const [error, setError] = useState<AiEditStreamError | null>(null);

  const controllerRef = useRef<AbortController | null>(null);
  // Identifies the active stream so out-of-order updates from a
  // previous start() (whose async work didn't notice the abort yet)
  // don't bleed into the new stream.
  const generationRef = useRef(0);

  const cancel = useCallback(() => {
    const controller = controllerRef.current;
    if (!controller) return;
    controller.abort();
    controllerRef.current = null;
  }, []);

  const reset = useCallback(() => {
    cancel();
    generationRef.current += 1;
    setStatus("idle");
    setReplacement("");
    setUsage(null);
    setError(null);
  }, [cancel]);

  // Cancel any in-flight stream when the consumer unmounts so we
  // don't leak a fetch + reader pair past the popover closing.
  useEffect(() => {
    return () => {
      controllerRef.current?.abort();
      controllerRef.current = null;
    };
  }, []);

  const start = useCallback(
    async (params: AiEditStartParams) => {
      controllerRef.current?.abort();
      const controller = new AbortController();
      controllerRef.current = controller;
      const generation = ++generationRef.current;

      setStatus("streaming");
      setReplacement("");
      setUsage(null);
      setError(null);

      let token: string | null = null;
      try {
        token = await getToken();
      } catch {
        token = null;
      }
      if (generationRef.current !== generation) return;

      const body: ApiSchemas["AIEditRequest"] = {
        selection_text: params.selectionText,
        selection_start: params.selectionStart,
        selection_end: params.selectionEnd,
        instruction: params.instruction,
        ...(params.docContext ? { doc_context: params.docContext } : {}),
      };

      let response: Response;
      try {
        response = await fetch(`${API_BASE}/study-guides/${guideId}/ai/edit`, {
          method: "POST",
          credentials: "same-origin",
          headers: {
            "Content-Type": "application/json",
            Accept: "text/event-stream",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify(body),
          signal: controller.signal,
        });
      } catch (err) {
        if (generationRef.current !== generation) return;
        if (controller.signal.aborted) {
          setStatus("cancelled");
          return;
        }
        setStatus("error");
        setError({ message: errorMessage(err) });
        return;
      }

      if (generationRef.current !== generation) return;

      if (!response.ok || !response.body) {
        setStatus("error");
        setError({
          message: await readErrorMessage(response),
          status: response.status,
        });
        return;
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";
      let receivedUsage: AiEditStreamUsage | null = null;

      try {
        while (true) {
          const { value, done } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });

          let separator = findSeparator(buffer);
          while (separator !== -1) {
            const frame = buffer.slice(0, separator);
            buffer = buffer.slice(
              separator + separatorLength(buffer, separator),
            );
            const event = parseFrame(frame);

            if (generationRef.current !== generation) return;

            if (event.type === "delta") {
              if (event.text) {
                setReplacement((prev) => prev + event.text);
              }
            } else if (event.type === "usage") {
              receivedUsage = event.usage;
              setUsage(event.usage);
            } else if (event.type === "error") {
              setStatus("error");
              setError({ message: event.message ?? "stream error" });
              controller.abort();
              return;
            } else if (event.type === "done") {
              setStatus("done");
              if (receivedUsage) setUsage(receivedUsage);
              controllerRef.current = null;
              return;
            }
            separator = findSeparator(buffer);
          }
        }

        if (generationRef.current !== generation) return;
        // Reader closed cleanly without a `done` event. The Go server
        // always emits `done` on success, so this is an abnormal EOF
        // (proxy timed out, network hiccup) -- surface it.
        setStatus("error");
        setError({ message: "Stream ended unexpectedly" });
      } catch (err) {
        if (generationRef.current !== generation) return;
        if (controller.signal.aborted) {
          setStatus("cancelled");
          return;
        }
        setStatus("error");
        setError({ message: errorMessage(err) });
      } finally {
        if (controllerRef.current === controller) {
          controllerRef.current = null;
        }
      }
    },
    [getToken, guideId],
  );

  return { status, replacement, usage, error, start, cancel, reset };
}

interface ParsedEvent {
  type: "delta" | "usage" | "error" | "done" | "unknown";
  text?: string;
  usage: AiEditStreamUsage;
  message?: string;
}

function parseFrame(frame: string): ParsedEvent {
  const result: ParsedEvent = {
    type: "unknown",
    usage: emptyUsage(),
  };
  const dataLines: string[] = [];

  for (const rawLine of frame.split(NEWLINE)) {
    if (!rawLine || rawLine.startsWith(":")) continue;
    const colon = rawLine.indexOf(":");
    if (colon === -1) continue;
    const field = rawLine.slice(0, colon);
    const value =
      rawLine.charAt(colon + 1) === " "
        ? rawLine.slice(colon + 2)
        : rawLine.slice(colon + 1);
    if (field === "event") {
      if (
        value === "delta" ||
        value === "usage" ||
        value === "error" ||
        value === "done"
      ) {
        result.type = value;
      }
    } else if (field === "data") {
      dataLines.push(value);
    }
  }

  if (dataLines.length === 0) return result;
  const data = dataLines.join("\n");
  try {
    const parsed: unknown = JSON.parse(data);
    if (result.type === "delta" && isObject(parsed)) {
      const delta = parsed as DeltaPayload;
      result.text = typeof delta.text === "string" ? delta.text : "";
    } else if (result.type === "usage" && isObject(parsed)) {
      const usage = parsed as UsagePayload;
      result.usage = {
        inputTokens: usage.input_tokens ?? 0,
        outputTokens: usage.output_tokens ?? 0,
        cacheReadTokens: usage.cache_read_tokens ?? 0,
        cacheWriteTokens: usage.cache_write_tokens ?? 0,
      };
    } else if (result.type === "error" && isObject(parsed)) {
      const err = parsed as ErrorPayload;
      result.message = typeof err.message === "string" ? err.message : "";
    }
  } catch {
    // malformed JSON -- treat as unknown
  }
  return result;
}

function emptyUsage(): AiEditStreamUsage {
  return {
    inputTokens: 0,
    outputTokens: 0,
    cacheReadTokens: 0,
    cacheWriteTokens: 0,
  };
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null;
}

function findSeparator(buf: string): number {
  const lf = buf.indexOf("\n\n");
  const crlf = buf.indexOf("\r\n\r\n");
  if (lf === -1) return crlf;
  if (crlf === -1) return lf;
  return Math.min(lf, crlf);
}

function separatorLength(buf: string, idx: number): number {
  return buf.startsWith("\r\n\r\n", idx) ? 4 : 2;
}

function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  return "Unexpected error";
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const body: unknown = await response.clone().json();
    if (
      isObject(body) &&
      "message" in body &&
      typeof body.message === "string"
    ) {
      return body.message;
    }
  } catch {
    // fall through
  }
  return `Request failed (${response.status})`;
}
