import { act, renderHook } from "@testing-library/react";

import { useDebouncedValue } from "./use-debounced-value";

describe("useDebouncedValue", () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("returns the initial value synchronously on first render", () => {
    const { result } = renderHook(() => useDebouncedValue("hello", 250));
    expect(result.current).toBe("hello");
  });

  it("only commits the latest value after the delay elapses", () => {
    const { result, rerender } = renderHook(
      ({ value }: { value: string }) => useDebouncedValue(value, 250),
      { initialProps: { value: "a" } },
    );

    rerender({ value: "ab" });
    rerender({ value: "abc" });
    expect(result.current).toBe("a");

    act(() => {
      jest.advanceTimersByTime(249);
    });
    expect(result.current).toBe("a");

    act(() => {
      jest.advanceTimersByTime(1);
    });
    expect(result.current).toBe("abc");
  });

  it("resets the timer on every change so rapid input never flushes early", () => {
    const { result, rerender } = renderHook(
      ({ value }: { value: number }) => useDebouncedValue(value, 100),
      { initialProps: { value: 0 } },
    );

    rerender({ value: 1 });
    act(() => {
      jest.advanceTimersByTime(80);
    });
    rerender({ value: 2 });
    act(() => {
      jest.advanceTimersByTime(80);
    });
    expect(result.current).toBe(0);

    act(() => {
      jest.advanceTimersByTime(20);
    });
    expect(result.current).toBe(2);
  });
});
