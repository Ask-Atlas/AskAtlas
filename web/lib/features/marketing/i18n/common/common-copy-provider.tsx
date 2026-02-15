"use client";

import React, { createContext, useContext } from "react";
import type { MarketingCommonDictionary } from "./types";
import enMarketingCommonDictionary from "./dictionaries/en";

const CommonCopyContext = createContext<MarketingCommonDictionary>(
  enMarketingCommonDictionary,
);

interface CommonCopyProviderProps {
  copy: MarketingCommonDictionary;
  children: React.ReactNode;
}

export function CommonCopyProvider({
  copy,
  children,
}: CommonCopyProviderProps) {
  return (
    <CommonCopyContext.Provider value={copy}>
      {children}
    </CommonCopyContext.Provider>
  );
}

export function useCommonCopy() {
  return useContext(CommonCopyContext);
}
