"use client";

import React, { createContext, useContext } from "react";
import type { DashboardCommonDictionary } from "./types";
import enDashboardCommonDictionary from "./dictionaries/en";

const DashboardCommonCopyContext = createContext<DashboardCommonDictionary>(
  enDashboardCommonDictionary,
);

interface DashboardCommonCopyProviderProps {
  copy: DashboardCommonDictionary;
  children: React.ReactNode;
}

export function DashboardCommonCopyProvider({
  copy,
  children,
}: DashboardCommonCopyProviderProps) {
  return (
    <DashboardCommonCopyContext.Provider value={copy}>
      {children}
    </DashboardCommonCopyContext.Provider>
  );
}

export function useDashboardCommonCopy() {
  return useContext(DashboardCommonCopyContext);
}
