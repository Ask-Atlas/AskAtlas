/**
 * Storybook-only stub for `@clerk/nextjs`. Clerk's React hooks throw
 * outside a real <ClerkProvider>, which would break every story that
 * indirectly mounts a component using `useAuth` / `useUser` / etc.
 *
 * Vite aliases `@clerk/nextjs` to this module via `.storybook/main.ts`
 * (`viteFinal`). It only ships in the storybook bundle -- production
 * builds and tests untouched.
 */

import type { ReactNode } from "react";

export function useAuth() {
  return {
    isLoaded: true,
    isSignedIn: false as boolean | undefined,
    userId: null as string | null,
    sessionId: null as string | null,
    getToken: async (): Promise<string | null> => null,
    signOut: async () => undefined,
  };
}

export function useUser() {
  return {
    isLoaded: true,
    isSignedIn: false as boolean | undefined,
    user: null as unknown,
  };
}

export function useClerk() {
  return {
    signOut: async () => undefined,
    openSignIn: () => undefined,
    openSignUp: () => undefined,
  };
}

export function ClerkProvider({ children }: { children: ReactNode }) {
  return <>{children}</>;
}

export function SignedIn(_props: { children?: ReactNode }) {
  return null;
}

export function SignedOut({ children }: { children?: ReactNode }) {
  return <>{children}</>;
}

export function SignInButton({ children }: { children?: ReactNode }) {
  return <>{children ?? <button type="button">Sign in</button>}</>;
}

export function SignUpButton({ children }: { children?: ReactNode }) {
  return <>{children ?? <button type="button">Sign up</button>}</>;
}

export function UserButton() {
  return null;
}
