import { act, render, screen, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom";

// The default resolver transitively imports the Clerk server SDK
// (ESM) via the resolveRefs server action. All tests here pass a
// custom resolver, so stubbing the module avoids pulling Clerk into
// jest.
jest.mock("../../../../api/actions/refs", () => ({
  resolveRefs: jest.fn(),
}));

import type { RefSummary } from "@/lib/api/types";

import {
  EntityRefProvider,
  useEntityRef,
  type RefResolver,
} from "./entity-ref-context";
import type { EntityRef } from "./extract-refs";

const SG = "11111111-2222-3333-4444-555555555555";
const QZ = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee";

function Probe({ type, id }: { type: "sg" | "quiz"; id: string }) {
  const { summary, status } = useEntityRef(type, id);
  return (
    <div>
      <span data-testid="status">{status}</span>
      <span data-testid="summary">{summary ? (summary.title ?? "") : "null"}</span>
    </div>
  );
}

describe("EntityRefProvider", () => {
  it("resolves refs on mount and exposes them via useEntityRef", async () => {
    const resolver: RefResolver = jest.fn().mockResolvedValue({
      [`sg:${SG}`]: {
        type: "sg",
        id: SG,
        title: "BST primer",
      } as RefSummary,
    });
    render(
      <EntityRefProvider refs={[{ type: "sg", id: SG }]} resolver={resolver}>
        <Probe type="sg" id={SG} />
      </EntityRefProvider>,
    );
    expect(screen.getByTestId("status")).toHaveTextContent("pending");
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );
    expect(screen.getByTestId("summary")).toHaveTextContent("BST primer");
  });

  it("uses initial data synchronously without calling resolver", () => {
    const resolver: RefResolver = jest.fn();
    render(
      <EntityRefProvider
        refs={[{ type: "sg", id: SG }]}
        initial={{
          [`sg:${SG}`]: {
            type: "sg",
            id: SG,
            title: "prefetched",
          } as RefSummary,
        }}
        resolver={resolver}
      >
        <Probe type="sg" id={SG} />
      </EntityRefProvider>,
    );
    expect(screen.getByTestId("status")).toHaveTextContent("ready");
    expect(screen.getByTestId("summary")).toHaveTextContent("prefetched");
    expect(resolver).not.toHaveBeenCalled();
  });

  it("invokes resolver exactly once even with duplicate refs", async () => {
    const resolver: RefResolver = jest.fn().mockResolvedValue({});
    const refs: EntityRef[] = [
      { type: "sg", id: SG },
      { type: "sg", id: SG },
      { type: "quiz", id: QZ },
    ];
    render(
      <EntityRefProvider refs={refs} resolver={resolver}>
        <Probe type="sg" id={SG} />
      </EntityRefProvider>,
    );
    await waitFor(() => expect(resolver).toHaveBeenCalledTimes(1));
  });

  it("refs missing from the response resolve to null", async () => {
    const resolver: RefResolver = jest.fn().mockResolvedValue({});
    render(
      <EntityRefProvider refs={[{ type: "sg", id: SG }]} resolver={resolver}>
        <Probe type="sg" id={SG} />
      </EntityRefProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );
    expect(screen.getByTestId("summary")).toHaveTextContent("null");
  });

  it("swallows resolver errors and settles into ready with no data", async () => {
    const resolver: RefResolver = jest
      .fn()
      .mockRejectedValue(new Error("network"));
    await act(async () => {
      render(
        <EntityRefProvider refs={[{ type: "sg", id: SG }]} resolver={resolver}>
          <Probe type="sg" id={SG} />
        </EntityRefProvider>,
      );
    });
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );
    expect(screen.getByTestId("summary")).toHaveTextContent("null");
  });

  it("empty refs list resolves to ready without calling resolver", () => {
    const resolver: RefResolver = jest.fn();
    render(
      <EntityRefProvider refs={[]} resolver={resolver}>
        <Probe type="sg" id={SG} />
      </EntityRefProvider>,
    );
    expect(screen.getByTestId("status")).toHaveTextContent("ready");
    expect(resolver).not.toHaveBeenCalled();
  });
});
