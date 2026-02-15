"use client";

import Link from "next/link";
import { Separator } from "@/components/ui/separator";
import { useCommonCopy } from "./i18n/common/common-copy-provider";

export function Footer() {
  const commonCopy = useCommonCopy();

  return (
    <footer className="w-full bg-background">
      <div className="mx-auto px-8 py-16">
        {/* Navigation */}
        <nav className="flex flex-wrap items-center justify-center gap-x-8 gap-y-3">
          {commonCopy.footer.links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
            >
              {link.label}
            </Link>
          ))}

          {commonCopy.footer.socialLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
            >
              {link.label}
              <span className="text-xs">↗</span>
            </a>
          ))}
        </nav>

        <Separator className="my-8 bg-border/50" />

        {/* Bottom row */}
        <div className="flex flex-col items-center gap-4 text-center">
          <Link
            href="#privacy"
            className="text-xs text-muted-foreground transition-colors hover:text-foreground"
          >
            {commonCopy.footer.privacyLabel}
          </Link>

          {/* Brand wordmark */}
          <p className="select-none text-[clamp(3rem,15vw,10rem)] font-bold leading-none tracking-tight bg-linear-to-r from-primary to-chart-2 bg-clip-text text-transparent">
            {commonCopy.footer.brandWordmark}
          </p>
        </div>
      </div>
    </footer>
  );
}
