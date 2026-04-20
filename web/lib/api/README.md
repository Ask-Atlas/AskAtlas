# `@/lib/api` — Typed API surface

Every call to the Go API lives here, fully typed end-to-end from the
OpenAPI spec. Frontend code never writes `fetch("/api/...")` by hand.

## Layout

```
lib/api/
├── generated/types.ts     # openapi-typescript output -- do not edit
├── server-client.ts       # openapi-fetch client + Clerk auth (server-only)
├── errors.ts              # ApiError, unwrap, unwrapVoid
├── types.ts               # named aliases for common schemas
├── client.ts              # legacy apiFetch wrapper (kept for back-compat)
├── index.ts               # barrel -> `import { ... } from "@/lib/api"`
└── actions/
    ├── files.ts
    ├── schools.ts
    ├── courses.ts
    ├── study-guides.ts
    ├── quizzes.ts
    ├── practice.ts
    └── me.ts
```

Each file in `actions/` starts with `"use server"`. Every exported
function is a Next.js Server Action: callable directly from Server
Components, and callable from Client Components as RPC without any
extra plumbing.

## Usage

### From a Server Component

```tsx
import { listDashboard } from "@/lib/api";

export default async function HomePage() {
  const dashboard = await listDashboard();
  return <DashboardView data={dashboard} />;
}
```

### From a Client Component

```tsx
"use client";
import { toggleStudyGuideFavorite } from "@/lib/api";

export function FavoriteButton({ id }: { id: string }) {
  return <button onClick={() => toggleStudyGuideFavorite(id)}>Toggle</button>;
}
```

### Error handling

All actions throw `ApiError` on non-2xx. The parsed `AppError`
envelope hangs off `.body`:

```ts
import { ApiError, getFile } from "@/lib/api";

try {
  const file = await getFile(id);
} catch (e) {
  if (e instanceof ApiError && e.status === 404) {
    notFound();
  }
  throw e;
}
```

## Auth

The server client wires a Clerk middleware that injects
`Authorization: Bearer <clerk-jwt>` per request. Callers never touch
tokens. When the request has no active session (e.g. a static build
fetching public content), the header is omitted and the Go API
returns its own 401 if the route requires auth.

## Configuration

Base URL resolution:

1. `API_BASE_URL` (server-only, e.g. `https://api-dev.askatlas.study/api`)
2. `NEXT_PUBLIC_API_BASE_URL`
3. `/api` (same-origin fallback for local dev via Next rewrite)

The OpenAPI spec declares `servers: - url: /api`, so the `/api`
prefix must already be in the resolved base URL.

## Regenerating types

```sh
make generate-api-types
```

Re-run after any `api/openapi.yaml` change. The checked-in file is
the source of truth the compiler reads; the script just wipes and
rewrites it deterministically.

## Testing

- `errors.test.ts` — pure unit tests for `unwrap` + `ApiError`.
- `actions.test.ts` — cross-domain integration tests. Mocks the
  `serverApi` client and asserts each action forwards the right
  `{ params, body }` shape into openapi-fetch's call site.
- `client.test.ts` — covers the legacy hand-rolled `apiFetch` wrapper.
  Kept until downstream callers migrate off it.

Run them with:

```sh
pnpm test -- lib/api
```

## Why this shape

We evaluated `openapi-fetch`, `orval`, and `hey-api`. openapi-fetch
wins on: zero codegen beyond types, 6kb runtime, middleware-based
auth, and same-author pairing with `openapi-typescript`. Server
Actions (rather than a React Query wrapper) keep Clerk auth + cache
invalidation on the server where Next.js already knows how to handle
them, and let Client Components call the same function as Server
Components without a separate hook layer.

See ASK-118 (type generation) + the PR that landed this module for
the full research trail.
