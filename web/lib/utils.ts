import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB"] as const;

/**
 * Human-readable byte size using binary (1024) units. Matches the UX
 * used by file pickers and cloud storage apps (52428800 -> "50 MB").
 */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return "0 B";
  }
  const exponent = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    BYTE_UNITS.length - 1,
  );
  const value = bytes / Math.pow(1024, exponent);
  // One decimal for sub-10 values, round otherwise -- keeps the badge tight.
  const rounded = value < 10 ? Math.round(value * 10) / 10 : Math.round(value);
  return `${rounded} ${BYTE_UNITS[exponent]}`;
}

const RELATIVE_THRESHOLDS: Array<{
  limit: number;
  divisor: number;
  unit: Intl.RelativeTimeFormatUnit;
}> = [
  { limit: 60, divisor: 1, unit: "second" },
  { limit: 3600, divisor: 60, unit: "minute" },
  { limit: 86_400, divisor: 3600, unit: "hour" },
  { limit: 604_800, divisor: 86_400, unit: "day" },
  { limit: 2_629_800, divisor: 604_800, unit: "week" },
  { limit: 31_557_600, divisor: 2_629_800, unit: "month" },
  { limit: Number.POSITIVE_INFINITY, divisor: 31_557_600, unit: "year" },
];

/**
 * Short relative timestamp backed by `Intl.RelativeTimeFormat` so the
 * copy follows the user's locale (e.g. "3d ago", "2 months ago"). Past
 * values only -- future timestamps collapse to "just now".
 */
export function formatRelativeDate(iso: string): string {
  const timestamp = Date.parse(iso);
  if (Number.isNaN(timestamp)) {
    return "";
  }
  const diffSeconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000));
  if (diffSeconds < 5) return "just now";

  const formatter = new Intl.RelativeTimeFormat(undefined, {
    numeric: "auto",
    style: "short",
  });
  for (const { limit, divisor, unit } of RELATIVE_THRESHOLDS) {
    if (diffSeconds < limit) {
      return formatter.format(-Math.floor(diffSeconds / divisor), unit);
    }
  }
  return formatter.format(-Math.floor(diffSeconds / 31_557_600), "year");
}
