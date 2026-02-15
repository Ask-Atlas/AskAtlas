import "server-only";
import { cookies, headers } from "next/headers";
import {
  hasLandingLocale,
  LANDING_DEFAULT_LOCALE,
  type LandingLocale,
} from "./get-landing-dictionary";

export async function resolveRequestLandingLocale(): Promise<LandingLocale> {
  const cookieStore = await cookies();
  const localeCookie = cookieStore.get("NEXT_LOCALE")?.value;

  if (localeCookie && hasLandingLocale(localeCookie)) {
    return localeCookie;
  }

  const headerStore = await headers();
  const acceptLanguage =
    headerStore.get("accept-language")?.toLowerCase() ?? "";

  if (acceptLanguage.includes("es")) {
    return "es";
  }

  return LANDING_DEFAULT_LOCALE;
}
