"use client";

/**
 * Lets a page tell the dashboard breadcrumb what to show for its
 * current segment. The default `DashboardBreadcrumb` only sees the
 * URL, so dynamic routes like `/courses/[courseId]` would otherwise
 * surface the raw UUID. Pages render `<SetDashboardBreadcrumb
 * label="CPTS 322" />` which writes into this context for the
 * lifetime of the route.
 */
import { createContext, useContext, useEffect, useMemo, useState } from "react";

interface BreadcrumbContextValue {
  currentLabel: string | null;
  setCurrentLabel: (label: string | null) => void;
}

const noop = () => undefined;

const DashboardBreadcrumbContext = createContext<BreadcrumbContextValue>({
  currentLabel: null,
  setCurrentLabel: noop,
});

export function DashboardBreadcrumbProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [currentLabel, setCurrentLabel] = useState<string | null>(null);
  const value = useMemo(
    () => ({ currentLabel, setCurrentLabel }),
    [currentLabel],
  );
  return (
    <DashboardBreadcrumbContext.Provider value={value}>
      {children}
    </DashboardBreadcrumbContext.Provider>
  );
}

export function useDashboardBreadcrumb() {
  return useContext(DashboardBreadcrumbContext);
}

/**
 * Drop this into a page (server or client) to override the trailing
 * breadcrumb label. Cleans up on unmount so a stale label never
 * leaks to a sibling route.
 */
export function SetDashboardBreadcrumb({ label }: { label: string }) {
  const { setCurrentLabel } = useDashboardBreadcrumb();
  useEffect(() => {
    setCurrentLabel(label);
    return () => setCurrentLabel(null);
  }, [label, setCurrentLabel]);
  return null;
}
