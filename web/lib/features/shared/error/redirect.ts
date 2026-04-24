/**
 * Full-page navigation helper for safety-net redirects (e.g. 401 token
 * expiry). Wrapped so tests can spy on redirect intent without fighting
 * jsdom's non-configurable `window.location`.
 */
export function hardRedirect(path: string): void {
  window.location.href = path;
}
