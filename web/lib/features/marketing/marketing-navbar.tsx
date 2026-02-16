"use client";
import {
  Navbar,
  NavBody,
  NavItems,
  MobileNav,
  NavbarLogo,
  NavbarButton,
  MobileNavHeader,
  MobileNavToggle,
  MobileNavMenu,
} from "@/components/ui/resizable-navbar";
import { ModeToggle } from "@/components/ui/mode-toggle";
import { SignedInButtons, SignedOutButtons } from "./auth-buttons";
import { useState } from "react";
import { useCommonCopy } from "./i18n/common/common-copy-provider";
import Link from "next/link";

export function MarketingNavbar() {
  const commonCopy = useCommonCopy();
  const navItems = commonCopy.nav.links.map((link) => ({
    name: link.label,
    link: link.href,
  }));
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  return (
    <div className="relative w-full">
      <Navbar>
        {/* Desktop Navigation */}
        <NavBody>
          <NavbarLogo />
          <NavItems items={navItems} />
          <div className="flex items-center gap-2">
            <NavbarButton as="div" variant="secondary" className="px-0 py-0">
              <ModeToggle />
            </NavbarButton>
            <SignedOutButtons />
            <SignedInButtons />
          </div>
        </NavBody>

        {/* Mobile Navigation */}
        <MobileNav>
          <MobileNavHeader>
            <NavbarLogo />
            <div className="flex items-center gap-2">
              <NavbarButton as="div" variant="secondary">
                <ModeToggle />
              </NavbarButton>
              <MobileNavToggle
                isOpen={isMobileMenuOpen}
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              />
            </div>
          </MobileNavHeader>

          <MobileNavMenu
            isOpen={isMobileMenuOpen}
            onClose={() => setIsMobileMenuOpen(false)}
          >
            {navItems.map((item) => (
              <Link
                key={item.link}
                href={item.link}
                onClick={() => setIsMobileMenuOpen(false)}
                className="relative text-neutral-600 dark:text-neutral-300"
                tabIndex={0}
                aria-label={item.name}
                role="link"
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    setIsMobileMenuOpen(false);
                  }
                }}
              >
                <span className="block">{item.name}</span>
              </Link>
            ))}
            <SignedOutButtons
              isMobile
              onClick={() => setIsMobileMenuOpen(false)}
            />
            <SignedInButtons
              isMobile
              onClick={() => setIsMobileMenuOpen(false)}
            />
          </MobileNavMenu>
        </MobileNav>
      </Navbar>
    </div>
  );
}
