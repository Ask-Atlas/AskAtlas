import "server-only";
import type { DashboardCommonDictionary } from "./types";

const dashboardCommonDictionaries = {
  en: () => import("./dictionaries/en").then((module) => module.default),
  es: () => import("./dictionaries/es").then((module) => module.default),
};

export type DashboardLocale = keyof typeof dashboardCommonDictionaries;

export const DASHBOARD_DEFAULT_LOCALE: DashboardLocale = "en";

export function hasDashboardLocale(locale: string): locale is DashboardLocale {
  return locale in dashboardCommonDictionaries;
}

export async function getDashboardCommonDictionary(
  locale: DashboardLocale,
): Promise<DashboardCommonDictionary> {
  return dashboardCommonDictionaries[locale]();
}
