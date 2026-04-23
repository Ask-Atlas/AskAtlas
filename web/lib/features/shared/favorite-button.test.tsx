/**
 * Covers the ASK-168 acceptance criteria for the primitive's local
 * contract. `useOptimistic` only exposes the flipped state during the
 * in-flight transition -- after the promise settles, state reverts to
 * `initialFavorited` so the caller can sync from the server-returned
 * `ToggleFavoriteResponse`. These tests verify that the click-time
 * UX is correct AND that `onToggle` fires exactly once per click.
 */
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import type { ToggleFavoriteResponse } from "@/lib/api/types";

import { FavoriteButton } from "./favorite-button";

function okResponse(favorited: boolean): ToggleFavoriteResponse {
  return {
    favorited,
    favorited_at: favorited ? "2026-04-23T10:00:00Z" : null,
  };
}

/**
 * Controllable deferred so the test can observe the optimistic state
 * WHILE `onToggle` is still pending.
 */
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe("FavoriteButton", () => {
  it("starts unpressed when initialFavorited is false", () => {
    render(
      <FavoriteButton
        initialFavorited={false}
        label="Favorite this file"
        onToggle={jest.fn().mockResolvedValue(okResponse(true))}
      />,
    );
    expect(screen.getByRole("button")).toHaveAttribute("aria-pressed", "false");
  });

  it("starts pressed when initialFavorited is true", () => {
    render(
      <FavoriteButton
        initialFavorited
        label="Favorite this file"
        onToggle={jest.fn().mockResolvedValue(okResponse(false))}
      />,
    );
    expect(screen.getByRole("button")).toHaveAttribute("aria-pressed", "true");
  });

  it("shows the optimistic state while onToggle is pending", async () => {
    const gate = deferred<ToggleFavoriteResponse>();
    const onToggle = jest.fn().mockReturnValue(gate.promise);

    render(
      <FavoriteButton
        initialFavorited={false}
        label="Favorite"
        onToggle={onToggle}
      />,
    );
    const button = screen.getByRole("button");

    await userEvent.click(button);
    await waitFor(() => expect(button).toHaveAttribute("aria-pressed", "true"));
    expect(onToggle).toHaveBeenCalledTimes(1);
    // Release the deferred so the transition can settle and we don't leak it.
    await act(async () => {
      gate.resolve(okResponse(true));
    });
  });

  it("reverts to initialFavorited once the transition settles (caller syncs from response)", async () => {
    const onToggle = jest.fn().mockResolvedValue(okResponse(true));
    render(
      <FavoriteButton
        initialFavorited={false}
        label="Favorite"
        onToggle={onToggle}
      />,
    );
    const button = screen.getByRole("button");
    await userEvent.click(button);
    await waitFor(() => expect(onToggle).toHaveBeenCalledTimes(1));
    // Parent hasn't re-rendered with a new initialFavorited, so React
    // commits back to the original value -- this is the expected contract.
    await waitFor(() =>
      expect(button).toHaveAttribute("aria-pressed", "false"),
    );
  });

  it("also reverts when onToggle rejects (rollback happens via the same settle path)", async () => {
    const onToggle = jest.fn().mockRejectedValue(new Error("network"));
    render(
      <FavoriteButton initialFavorited label="Favorite" onToggle={onToggle} />,
    );
    const button = screen.getByRole("button");

    await act(async () => {
      await userEvent.click(button);
    });

    await waitFor(() => expect(onToggle).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(button).toHaveAttribute("aria-pressed", "true"));
  });
});
