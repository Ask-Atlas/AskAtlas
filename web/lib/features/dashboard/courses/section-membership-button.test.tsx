/**
 * Exercises the ASK-182 acceptance criteria: three-state rendering
 * (unknown/not-member/member), optimistic join + revert on failure,
 * and the member-state leave flow via ConfirmationDialog (cancel
 * doesn't call onLeave, confirm does).
 */
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";

import { SectionMembershipButton } from "./section-membership-button";

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (err: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe("SectionMembershipButton / unknown state", () => {
  it("renders a disabled spinner while membership is being resolved", () => {
    render(
      <SectionMembershipButton
        membership="unknown"
        onJoin={jest.fn()}
        onLeave={jest.fn()}
      />,
    );
    const button = screen.getByRole("button", { name: /checking enrollment/i });
    expect(button).toBeDisabled();
  });
});

describe("SectionMembershipButton / not-member state", () => {
  it('renders a "Join" button', () => {
    render(
      <SectionMembershipButton
        membership="not-member"
        onJoin={jest.fn().mockResolvedValue(undefined)}
        onLeave={jest.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: "Join" })).toBeInTheDocument();
  });

  it("fires onJoin when clicked and keeps the button disabled while pending", async () => {
    const joinPromise = deferred<void>();
    const onJoin = jest.fn(() => joinPromise.promise);
    render(
      <SectionMembershipButton
        membership="not-member"
        onJoin={onJoin}
        onLeave={jest.fn()}
      />,
    );
    await userEvent.click(screen.getByRole("button"));
    expect(onJoin).toHaveBeenCalledTimes(1);
    // The button must not accept repeat clicks while the request is
    // in flight -- otherwise a double-click enrolls the user twice.
    expect(screen.getByRole("button")).toBeDisabled();
    joinPromise.resolve();
    await waitFor(() => expect(onJoin).toHaveBeenCalledTimes(1));
  });

  it("reverts to the Join label when onJoin rejects (AC3)", async () => {
    const joinPromise = deferred<void>();
    const onJoin = jest.fn(() => joinPromise.promise);
    render(
      <SectionMembershipButton
        membership="not-member"
        onJoin={onJoin}
        onLeave={jest.fn()}
      />,
    );
    await userEvent.click(screen.getByRole("button"));
    // Rejection is wrapped in `act` so React flushes the settled-transition
    // state update (useOptimistic revert) before our assertion runs.
    await act(async () => {
      joinPromise.reject(new Error("network"));
    });
    // After the transition settles, useOptimistic reverts and the
    // button returns to the pre-click label. Caller is responsible
    // for surfacing a toast via its own try/catch around onJoin.
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "Join" })).toBeInTheDocument(),
    );
  });
});

describe("SectionMembershipButton / member state", () => {
  it('renders an "Enrolled" button', () => {
    render(
      <SectionMembershipButton
        membership="member"
        onJoin={jest.fn()}
        onLeave={jest.fn().mockResolvedValue(undefined)}
      />,
    );
    expect(
      screen.getByRole("button", { name: "Enrolled" }),
    ).toBeInTheDocument();
  });

  it("opens the leave confirmation dialog without firing onLeave", async () => {
    const onLeave = jest.fn().mockResolvedValue(undefined);
    render(
      <SectionMembershipButton
        membership="member"
        onJoin={jest.fn()}
        onLeave={onLeave}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Enrolled" }));
    expect(
      screen.getByRole("alertdialog", { name: /leave this section/i }),
    ).toBeInTheDocument();
    expect(onLeave).not.toHaveBeenCalled();
  });

  it("does not fire onLeave when the dialog is cancelled", async () => {
    const onLeave = jest.fn().mockResolvedValue(undefined);
    render(
      <SectionMembershipButton
        membership="member"
        onJoin={jest.fn()}
        onLeave={onLeave}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Enrolled" }));
    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onLeave).not.toHaveBeenCalled();
  });

  it("fires onLeave when the dialog is confirmed (AC2)", async () => {
    const leavePromise = deferred<void>();
    const onLeave = jest.fn(() => leavePromise.promise);
    render(
      <SectionMembershipButton
        membership="member"
        onJoin={jest.fn()}
        onLeave={onLeave}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Enrolled" }));
    await userEvent.click(screen.getByRole("button", { name: "Leave" }));
    expect(onLeave).toHaveBeenCalledTimes(1);
    leavePromise.resolve();
  });
});
