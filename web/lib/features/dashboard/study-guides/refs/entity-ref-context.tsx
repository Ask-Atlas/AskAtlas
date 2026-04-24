"use client";

import { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";

import type { RefSummary } from "@/lib/api/types";

import { refKey, type EntityRef, type EntityType } from "./extract-refs";

type Status = "pending" | "ready";

interface ContextValue {
  status: Status;
  get(type: EntityType, id: string): RefSummary | null;
}

const EntityRefContext = createContext<ContextValue>({
  status: "ready",
  get: () => null,
});

export type RefResolver = (
  refs: EntityRef[],
) => Promise<Record<string, RefSummary | null | undefined>>;

interface EntityRefProviderProps {
  refs: EntityRef[];
  initial?: Record<string, RefSummary | null>;
  resolver?: RefResolver;
  children: ReactNode;
}

async function defaultResolver(
  refs: EntityRef[],
): Promise<Record<string, RefSummary | null | undefined>> {
  if (refs.length === 0) return {};
  const { resolveRefs } = await import("@/lib/api/actions/refs");
  const resp = await resolveRefs({ refs });
  return resp.results ?? {};
}

export function EntityRefProvider({
  refs,
  initial,
  resolver = defaultResolver,
  children,
}: EntityRefProviderProps) {
  const [map, setMap] = useState<Record<string, RefSummary | null>>(
    initial ?? {},
  );
  const [status, setStatus] = useState<Status>(
    initial || refs.length === 0 ? "ready" : "pending",
  );

  const refsKey = refs
    .map((r) => refKey(r.type, r.id))
    .sort()
    .join(",");

  useEffect(() => {
    if (initial) return;
    if (refs.length === 0) {
      setStatus("ready");
      return;
    }
    let cancelled = false;
    setStatus("pending");
    resolver(refs)
      .then((results) => {
        if (cancelled) return;
        const next: Record<string, RefSummary | null> = {};
        for (const [k, v] of Object.entries(results)) {
          next[k.toLowerCase()] = v ?? null;
        }
        setMap(next);
        setStatus("ready");
      })
      .catch(() => {
        if (cancelled) return;
        setMap({});
        setStatus("ready");
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refsKey, resolver]);

  return (
    <EntityRefContext.Provider
      value={{
        status,
        get: (type, id) => map[refKey(type, id)] ?? null,
      }}
    >
      {children}
    </EntityRefContext.Provider>
  );
}

export function useEntityRef(type: EntityType, id: string) {
  const ctx = useContext(EntityRefContext);
  return { summary: ctx.get(type, id), status: ctx.status };
}
