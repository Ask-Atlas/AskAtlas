---
sidebar_position: 3
---

# Clerk Setup

This guide covers how to configure [Clerk](https://clerk.com) for AskAtlas. We use separate Clerk applications per environment.

## Clerk Applications

| Environment | Clerk App |
|-------------|-----------|
| Dev | Separate Clerk app (dev keys) |
| Stage | Separate Clerk app (staging keys) |
| Prod | Not yet provisioned |

Each app has its own API keys and webhook configuration.

## Creating a Clerk Application

1. Go to [clerk.com/dashboard](https://dashboard.clerk.com/) and sign in
2. Click **Create application**
3. Name it something clear (e.g. `AskAtlas Dev`, `AskAtlas Stage`)
4. Choose your sign-in methods (email, Google, etc.)
5. Copy the keys from the dashboard

## Getting Your Keys

From the Clerk dashboard, navigate to **API Keys**. You'll need:

| Key | Where It's Used | Environment Variable |
|-----|-----------------|---------------------|
| **Publishable Key** | Frontend (Next.js) | `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` |
| **Secret Key** | Backend (Go API) | `CLERK_SECRET_KEY` |

Both keys are managed through Infisical. Add them to the appropriate Infisical environment so they're injected at runtime.

## Configuring Webhooks

Clerk sends webhook events when users are created, updated, or deleted. Our Go API listens for these events to keep the local database in sync.

### Steps

1. In the Clerk dashboard, navigate to **Webhooks**
2. Click **Add Endpoint**
3. Set the endpoint URL to your API's webhook route:
   ```
   https://<your-api-domain>/webhooks/clerk
   ```
   Replace `<your-api-domain>` with the appropriate environment's API domain.
4. Subscribe to the following events:
   - `user.created`
   - `user.updated`
   - `user.deleted`
5. Click **Create**
6. Copy the **Signing Secret** — this is your `CLERK_WEBHOOK_SECRET`
7. Add `CLERK_WEBHOOK_SECRET` to Infisical for the matching environment

### Verifying Webhooks Work

After configuring the webhook:

1. Create a test user via the Clerk dashboard or sign-up flow
2. Check the Go API logs for webhook receipt
3. Verify the user appears in your PostgreSQL `users` table
4. Check the Clerk dashboard **Webhooks → Logs** tab for delivery status

## Registering the Clerk SDK (Go API)

The Go API registers the Clerk secret key at startup in `cmd/api/main.go`:

```go
clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
```

This must be called before any middleware or handlers that use the Clerk SDK.

## Frontend Integration

The Next.js app uses the `@clerk/nextjs` package. Key files:

| File | Purpose |
|------|---------|
| `app/layout.tsx` | Wraps the app in `<ClerkProvider>` |
| `proxy.ts` | Clerk middleware protecting dashboard routes |
| `components/providers.tsx` | Theme and auth providers |

### Route Protection

Protected routes are configured in `proxy.ts`:

```typescript
const isDashboardRoute = createRouteMatcher([
  "/home(.*)",
  "/courses(.*)",
  "/resources(.*)",
  "/study-guides(.*)",
  "/me(.*)",
]);

export default clerkMiddleware(async (auth, req) => {
  if (isDashboardRoute(req)) {
    await auth.protect();
  }
});
```

To protect a new route, add its pattern to the `createRouteMatcher` array.

## Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| 401 on API calls | Missing or invalid JWT | Check `CLERK_SECRET_KEY` is set correctly in Infisical |
| Webhook events not received | Wrong endpoint URL | Verify the URL in Clerk dashboard matches your API domain |
| Webhook signature verification fails | Wrong secret | Ensure `CLERK_WEBHOOK_SECRET` matches the signing secret from Clerk |
| User created in Clerk but not in DB | Webhook not processed | Check API logs and Clerk webhook delivery logs |
