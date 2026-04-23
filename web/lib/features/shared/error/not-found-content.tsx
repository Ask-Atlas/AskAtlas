import Link from "next/link";

import { Button } from "@/components/ui/button";

export function NotFoundContent() {
  return (
    <main className="flex min-h-[60vh] flex-col items-center justify-center gap-4 p-8">
      <h1 className="text-2xl font-semibold">Page not found</h1>
      <p className="text-muted-foreground max-w-md text-center">
        This page doesn&apos;t exist or has moved.
      </p>
      <Button asChild>
        <Link href="/home">Back to dashboard</Link>
      </Button>
    </main>
  );
}
