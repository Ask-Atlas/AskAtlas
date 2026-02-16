import "server-only";

import { cookies, headers } from "next/headers";
import {
  hasDashboardLocale,
  DASHBOARD_DEFAULT_LOCALE,
  type DashboardLocale,
} from "./get-common-dictionary";

export async function resolveRequestDashboardLocale(): Promise<DashboardLocale> {
  const cookieStore = await cookies();
  const localeCookie = cookieStore.get("NEXT_LOCALE")?.value;

  if (localeCookie && hasDashboardLocale(localeCookie)) {
    return localeCookie;
  }

  const headerStore = await headers();
  const acceptLanguage =
    headerStore.get("accept-language")?.toLowerCase() ?? "";

  if (acceptLanguage.includes("es")) {
    return "es";
  }

  return DASHBOARD_DEFAULT_LOCALE;
}
