"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/errors";

import { hardRedirect } from "./redirect";

interface ErrorBoundaryContentProps {
  error: Error & { digest?: string };
  reset: () => void;
}

const SIGN_IN_PATH = "/sign-in";

export function ErrorBoundaryContent({
  error,
  reset,
}: ErrorBoundaryContentProps) {
  const isAuthExpired = error instanceof ApiError && error.status === 401;

  useEffect(() => {
    if (isAuthExpired) {
      hardRedirect(SIGN_IN_PATH);
    }
  }, [isAuthExpired]);

  if (isAuthExpired) {
    return null;
  }

  const message =
    error instanceof ApiError && error.body?.message
      ? error.body.message
      : "We hit an unexpected problem.";

  return (
    <main className="flex min-h-[60vh] flex-col items-center justify-center gap-4 p-8">
      <h1 className="text-2xl font-semibold">Something went wrong</h1>
      <p className="text-muted-foreground max-w-md text-center">{message}</p>
      <Button onClick={reset}>Try again</Button>
    </main>
  );
}
