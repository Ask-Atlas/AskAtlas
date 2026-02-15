"use client";

import React, { createContext, useContext } from "react";
import type { LandingDictionary } from "./types";
import enLandingDictionary from "./dictionaries/en";

const LandingCopyContext =
  createContext<LandingDictionary>(enLandingDictionary);

interface LandingCopyProviderProps {
  copy: LandingDictionary;
  children: React.ReactNode;
}

export function LandingCopyProvider({
  copy,
  children,
}: LandingCopyProviderProps) {
  return (
    <LandingCopyContext.Provider value={copy}>
      {children}
    </LandingCopyContext.Provider>
  );
}

export function useLandingCopy() {
  return useContext(LandingCopyContext);
}
