import { auth } from "@clerk/nextjs/server";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { NotFoundContent } from "@/lib/features/shared/error/not-found-content";

export default async function RootNotFound() {
  const { userId } = await auth();
  const isAuthenticated = Boolean(userId);

  return (
    <NotFoundContent
      action={
        <Button asChild>
          <Link href={isAuthenticated ? "/home" : "/"}>
            {isAuthenticated ? "Back to dashboard" : "Back to home"}
          </Link>
        </Button>
      }
    />
  );
}
