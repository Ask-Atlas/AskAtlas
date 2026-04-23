"use client";

import { ErrorBoundaryContent } from "@/lib/features/shared/error/error-boundary-content";

export default function RootError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return <ErrorBoundaryContent error={error} reset={reset} />;
}
