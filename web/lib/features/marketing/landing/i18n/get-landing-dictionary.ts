import "server-only";
import type { LandingDictionary } from "./types";

const landingDictionaries = {
  en: () => import("./dictionaries/en").then((module) => module.default),
  es: () => import("./dictionaries/es").then((module) => module.default),
};

export type LandingLocale = keyof typeof landingDictionaries;

export const LANDING_DEFAULT_LOCALE: LandingLocale = "en";

export const LANDING_SUPPORTED_LOCALES: LandingLocale[] = Object.keys(
  landingDictionaries,
) as LandingLocale[];

export function hasLandingLocale(locale: string): locale is LandingLocale {
  return locale in landingDictionaries;
}

export async function getLandingDictionary(
  locale: LandingLocale,
): Promise<LandingDictionary> {
  return landingDictionaries[locale]();
}
