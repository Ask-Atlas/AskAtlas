/**
 * @jest-environment jsdom
 */
// jsdom doesn't ship TextEncoder/ReadableStream as globals. Pull them
// from Node's built-ins so the SSE-byte plumbing works under test.
// Response is duck-typed below so we don't need to polyfill it.
import { ReadableStream as NodeReadableStream } from "node:stream/web";
import {
  TextDecoder as NodeTextDecoder,
  TextEncoder as NodeTextEncoder,
} from "node:util";

/* eslint-disable @typescript-eslint/no-explicit-any */
if (typeof (globalThis as any).TextEncoder === "undefined") {
  (globalThis as any).TextEncoder = NodeTextEncoder;
}
if (typeof (globalThis as any).TextDecoder === "undefined") {
  (globalThis as any).TextDecoder = NodeTextDecoder;
}
if (typeof (globalThis as any).ReadableStream === "undefined") {
  (globalThis as any).ReadableStream = NodeReadableStream;
}
/* eslint-enable @typescript-eslint/no-explicit-any */

import { act, renderHook, waitFor } from "@testing-library/react";

import { useAiEditStream } from "./use-ai-edit-stream";

const mockGetToken = jest.fn<Promise<string | null>, []>();

jest.mock("@clerk/nextjs", () => ({
  useAuth: () => ({ getToken: mockGetToken }),
}));

const GUIDE_ID = "11111111-2222-3333-4444-555555555555";
const VALID_PARAMS = {
  selectionText: "hello",
  selectionStart: 0,
  selectionEnd: 5,
  instruction: "make it shorter",
};

interface FakeStreamHandle {
  push: (chunk: string) => void;
  close: () => void;
  error: (err: Error) => void;
}

function makeStream(): {
  body: ReadableStream<Uint8Array>;
  handle: FakeStreamHandle;
} {
  const encoder = new TextEncoder();
  let controllerRef: ReadableStreamDefaultController<Uint8Array> | null = null;
  const body = new ReadableStream<Uint8Array>({
    start(controller) {
      controllerRef = controller;
    },
  });
  const handle: FakeStreamHandle = {
    push: (chunk) => controllerRef?.enqueue(encoder.encode(chunk)),
    close: () => controllerRef?.close(),
    error: (err) => controllerRef?.error(err),
  };
  return { body, handle };
}

// Minimal duck-typed Response — the hook only reads `.ok`, `.body`,
// `.status`, and `.clone().json()`. We avoid the real Response
// constructor because jsdom + Node 20 in jest don't ship it.
function streamingResponse(body: ReadableStream<Uint8Array>): unknown {
  return {
    ok: true,
    status: 200,
    body,
    clone() {
      return { json: async () => ({}) };
    },
  };
}

function jsonErrorResponse(status: number, payload: unknown): unknown {
  const cloned = { json: async () => payload };
  return {
    ok: false,
    status,
    body: null,
    clone() {
      return cloned;
    },
  };
}

describe("useAiEditStream", () => {
  beforeEach(() => {
    mockGetToken.mockReset();
    mockGetToken.mockResolvedValue("fake-token");
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("starts in idle state", () => {
    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));
    expect(result.current.status).toBe("idle");
    expect(result.current.replacement).toBe("");
    expect(result.current.error).toBeNull();
  });

  it("transitions through streaming → done and accumulates deltas", async () => {
    const { body, handle } = makeStream();
    (global.fetch as jest.Mock).mockResolvedValue(streamingResponse(body));

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start(VALID_PARAMS);
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    await act(async () => {
      handle.push('event: delta\ndata: {"text":"Hi"}\n\n');
    });
    await waitFor(() => expect(result.current.replacement).toBe("Hi"));

    await act(async () => {
      handle.push('event: delta\ndata: {"text":" there"}\n\n');
      handle.push(
        'event: usage\ndata: {"input_tokens":12,"output_tokens":3,"cache_read_tokens":0,"cache_write_tokens":0}\n\n',
      );
      handle.push("event: done\ndata: {}\n\n");
      handle.close();
    });

    await waitFor(() => expect(result.current.status).toBe("done"));
    expect(result.current.replacement).toBe("Hi there");
    expect(result.current.usage).toEqual({
      inputTokens: 12,
      outputTokens: 3,
      cacheReadTokens: 0,
      cacheWriteTokens: 0,
    });
  });

  it("sends Authorization header + correct request body", async () => {
    const { body, handle } = makeStream();
    const fetchMock = jest.fn().mockResolvedValue(streamingResponse(body));
    global.fetch = fetchMock;

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start({
        ...VALID_PARAMS,
        docContext: { title: "My Guide", preceding: "before" },
      });
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toContain(`/study-guides/${GUIDE_ID}/ai/edit`);
    expect(init.method).toBe("POST");
    expect(init.headers["Authorization"]).toBe("Bearer fake-token");
    expect(init.headers["Accept"]).toBe("text/event-stream");
    const sent = JSON.parse(init.body);
    expect(sent).toEqual({
      selection_text: "hello",
      selection_start: 0,
      selection_end: 5,
      instruction: "make it shorter",
      doc_context: { title: "My Guide", preceding: "before" },
    });

    await act(async () => {
      handle.push("event: done\ndata: {}\n\n");
      handle.close();
    });
  });

  it("surfaces a server error event", async () => {
    const { body, handle } = makeStream();
    (global.fetch as jest.Mock).mockResolvedValue(streamingResponse(body));

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start(VALID_PARAMS);
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    await act(async () => {
      handle.push('event: error\ndata: {"message":"upstream blew up"}\n\n');
      handle.close();
    });

    await waitFor(() => expect(result.current.status).toBe("error"));
    expect(result.current.error?.message).toBe("upstream blew up");
  });

  it("surfaces an HTTP 4xx response with the AppError message", async () => {
    (global.fetch as jest.Mock).mockResolvedValue(
      jsonErrorResponse(429, { message: "rate limited" }),
    );

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      await result.current.start(VALID_PARAMS);
    });

    expect(result.current.status).toBe("error");
    expect(result.current.error).toEqual({
      message: "rate limited",
      status: 429,
    });
  });

  it("transitions to cancelled when cancel() is called mid-stream", async () => {
    const { body, handle } = makeStream();
    (global.fetch as jest.Mock).mockResolvedValue(streamingResponse(body));

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start(VALID_PARAMS);
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    await act(async () => {
      handle.push('event: delta\ndata: {"text":"so"}\n\n');
    });
    await waitFor(() => expect(result.current.replacement).toBe("so"));

    await act(async () => {
      result.current.cancel();
      handle.error(new DOMException("aborted", "AbortError"));
    });

    await waitFor(() => expect(result.current.status).toBe("cancelled"));
  });

  it("treats EOF without a `done` event as a stream error", async () => {
    const { body, handle } = makeStream();
    (global.fetch as jest.Mock).mockResolvedValue(streamingResponse(body));

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start(VALID_PARAMS);
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    await act(async () => {
      handle.push('event: delta\ndata: {"text":"partial"}\n\n');
      handle.close();
    });

    await waitFor(() => expect(result.current.status).toBe("error"));
    expect(result.current.error?.message).toMatch(/unexpected/i);
  });

  it("ignores comment-only frames (heartbeats)", async () => {
    const { body, handle } = makeStream();
    (global.fetch as jest.Mock).mockResolvedValue(streamingResponse(body));

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start(VALID_PARAMS);
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    await act(async () => {
      handle.push(": heartbeat\n\n");
      handle.push('event: delta\ndata: {"text":"ok"}\n\n');
      handle.push("event: done\ndata: {}\n\n");
      handle.close();
    });

    await waitFor(() => expect(result.current.status).toBe("done"));
    expect(result.current.replacement).toBe("ok");
  });

  it("reset() clears state and aborts any in-flight stream", async () => {
    const { body, handle } = makeStream();
    (global.fetch as jest.Mock).mockResolvedValue(streamingResponse(body));

    const { result } = renderHook(() => useAiEditStream({ guideId: GUIDE_ID }));

    await act(async () => {
      void result.current.start(VALID_PARAMS);
    });
    await waitFor(() => expect(result.current.status).toBe("streaming"));

    await act(async () => {
      handle.push('event: delta\ndata: {"text":"x"}\n\n');
    });
    await waitFor(() => expect(result.current.replacement).toBe("x"));

    await act(async () => {
      result.current.reset();
    });

    expect(result.current.status).toBe("idle");
    expect(result.current.replacement).toBe("");
    expect(result.current.error).toBeNull();
  });
});
