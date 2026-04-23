import Link from "next/link";

import { Button } from "@/components/ui/button";
import { NotFoundContent } from "@/lib/features/shared/error/not-found-content";

export default function DashboardNotFound() {
  return (
    <NotFoundContent
      action={
        <Button asChild>
          <Link href="/home">Back to dashboard</Link>
        </Button>
      }
    />
  );
}
