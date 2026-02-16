import { NavbarButton } from "@/components/ui/resizable-navbar";
import {
  SignInButton,
  SignUpButton,
  SignedOut,
  SignedIn,
  UserButton,
} from "@clerk/nextjs";
import { useCommonCopy } from "./i18n/common/common-copy-provider";

interface AuthButtonsProps {
  onClick?: () => void;
  isMobile?: boolean;
}

export function SignedOutButtons({ onClick, isMobile }: AuthButtonsProps) {
  const commonCopy = useCommonCopy();
  return (
    <SignedOut>
      <SignInButton mode="modal" forceRedirectUrl="/home">
        <NavbarButton
          as="button"
          type="button"
          variant="secondary"
          className={isMobile ? "w-full" : undefined}
          aria-label={commonCopy.nav.loginCta}
          onClick={onClick}
        >
          {commonCopy.nav.loginCta}
        </NavbarButton>
      </SignInButton>
      <SignUpButton mode="modal" forceRedirectUrl="/home">
        <NavbarButton
          as="button"
          type="button"
          variant="primary"
          className={isMobile ? "w-full" : undefined}
          aria-label={commonCopy.nav.primaryCta}
          onClick={onClick}
        >
          {commonCopy.nav.primaryCta}
        </NavbarButton>
      </SignUpButton>
    </SignedOut>
  );
}

export function SignedInButtons({ onClick, isMobile }: AuthButtonsProps) {
  const commonCopy = useCommonCopy();
  return (
    <SignedIn>
      <NavbarButton
        href="/home"
        variant="secondary"
        className={isMobile ? "w-full" : undefined}
        aria-label={commonCopy.nav.dashboardCta}
        onClick={onClick}
      >
        {commonCopy.nav.dashboardCta}
      </NavbarButton>
      <span className={isMobile ? "flex justify-center pt-2" : undefined}>
        <UserButton />
      </span>
    </SignedIn>
  );
}
