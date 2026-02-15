import "server-only";
import type { LandingLocale } from "@/lib/features/marketing/landing/i18n/get-landing-dictionary";
import type { MarketingCommonDictionary } from "./types";

const commonDictionaries = {
  en: () => import("./dictionaries/en").then((module) => module.default),
  es: () => import("./dictionaries/es").then((module) => module.default),
};

export async function getMarketingCommonDictionary(
  locale: LandingLocale,
): Promise<MarketingCommonDictionary> {
  return commonDictionaries[locale]();
}
