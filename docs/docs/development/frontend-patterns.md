---
sidebar_position: 2
---

# Frontend Patterns

This guide covers the conventions and patterns used in the Next.js frontend.

## Server-Driven Pages

**Pages should be server components.** They are thin orchestrators that compose feature components — they should not contain business logic or complex markup.

### Reference: Marketing Page

The marketing page is the gold standard for this pattern:

```tsx
// app/(marketing)/page.tsx
import { Hero } from "@/lib/features/marketing/landing/hero";
import { FeaturesDisplay } from "@/lib/features/marketing/landing/features-display";
import { Mission } from "@/lib/features/marketing/landing/mission";
import { CTA } from "@/lib/features/marketing/landing/cta";

export default function MarketingPage() {
  return (
    <>
      <Hero />
      <FeaturesDisplay />
      <Mission />
      <CTA />
    </>
  );
}
```

The page does nothing but compose feature components. All logic, styling, and content lives in the feature modules.

## Layouts Handle Data Fetching

Server-side data fetching (i18n dictionaries, auth context, etc.) happens in **layouts**, not pages. Layouts wrap children in context providers.

### Reference: Marketing Layout

```tsx
// app/(marketing)/layout.tsx
export default async function MarketingLayout({ children }) {
  const locale = await resolveRequestLandingLocale();
  const [landingCopy, commonCopy] = await Promise.all([
    getLandingDictionary(locale),
    getMarketingCommonDictionary(locale),
  ]);

  return (
    <CommonCopyProvider copy={commonCopy}>
      <LandingCopyProvider copy={landingCopy}>
        <header><MarketingNavbar /></header>
        <main>{children}</main>
        <Footer />
      </LandingCopyProvider>
    </CommonCopyProvider>
  );
}
```

Key patterns:
- Layout is an `async` server component — does the data fetching
- Dictionaries are resolved server-side, not in client components
- Providers wrap children so feature components can access copy via hooks

## Component Colocation

**Components live in `lib/features/<feature>/`, NOT in `app/`.** Pages import from features.

```
lib/features/
├── marketing/
│   ├── landing/
│   │   ├── hero.tsx
│   │   ├── features-display.tsx
│   │   ├── mission.tsx
│   │   ├── cta.tsx
│   │   └── i18n/
│   │       ├── landing-copy-provider.tsx
│   │       ├── get-landing-dictionary.ts
│   │       └── resolve-request-locale.ts
│   ├── marketing-navbar.tsx
│   ├── footer.tsx
│   └── section.tsx
└── dashboard/
    ├── sidebar.tsx
    └── i18n/
        └── ...
```

### Rules

- **`app/` contains only pages, layouts, and route configuration** — no components or business logic
- **`lib/features/` is organized by feature domain** — marketing, dashboard, etc.
- **Feature-scoped components stay in their feature** — `landing/hero.tsx` doesn't go in a generic `components/` folder
- **Shared UI primitives** (buttons, inputs, etc.) live in `components/ui/` (shadcn/ui) and `components/animate-ui/`

## Route Groups

Next.js route groups (parenthesized names) provide layout boundaries without affecting the URL:

```
app/
├── (marketing)/     → Public pages (landing, about)
│   ├── layout.tsx   → Marketing navbar, footer, i18n
│   └── page.tsx     → Landing page (/)
├── (dashboard)/     → Authenticated pages
│   ├── layout.tsx   → Sidebar, breadcrumbs, i18n
│   ├── home/        → /home
│   ├── courses/     → /courses
│   ├── resources/   → /resources
│   ├── study-guides/ → /study-guides
│   └── me/          → /me
└── practice/        → /practice (no group — standalone layout)
```

### When to Use Route Groups

- **Different layouts** — Marketing pages have navbar + footer. Dashboard has sidebar + breadcrumbs.
- **Different auth requirements** — Dashboard routes are protected, marketing routes are public.
- **Shared data fetching** — Each group's layout can fetch different data server-side.

## i18n

Internationalization uses server-side dictionary resolution with React context providers:

1. **Layout** resolves the locale from the request
2. **Layout** fetches the dictionary for that locale
3. **Layout** wraps children in a `CopyProvider`
4. **Feature components** use a `useCopy()` hook to access translations

Each feature has its own i18n directory with locale-specific dictionaries.

## Shared UI Components

| Directory | Purpose |
|-----------|---------|
| `components/ui/` | shadcn/ui primitives (Button, Separator, Sidebar, Tooltip, etc.) |
| `components/animate-ui/` | Animation components |
| `components/providers.tsx` | App-wide providers (Clerk, theme) |

These are feature-agnostic and shared across all pages.

## Adding a New Page

1. **Create the page** in `app/(<group>)/<route>/page.tsx` — keep it thin (server component)
2. **Create feature components** in `lib/features/<feature>/` — all UI and logic goes here
3. **Add i18n** if needed — create dictionaries in `lib/features/<feature>/i18n/`
4. **Update the layout** if the page needs data fetch or new providers
5. **Protect the route** — add the path to `proxy.ts` if it requires authentication
