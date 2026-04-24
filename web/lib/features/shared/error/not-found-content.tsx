import type { ReactNode } from "react";

interface NotFoundContentProps {
  action: ReactNode;
}

export function NotFoundContent({ action }: NotFoundContentProps) {
  return (
    <main className="flex min-h-[60vh] flex-col items-center justify-center gap-4 p-8">
      <h1 className="text-2xl font-semibold">Page not found</h1>
      <p className="text-muted-foreground max-w-md text-center">
        This page doesn&apos;t exist or has moved.
      </p>
      {action}
    </main>
  );
}
