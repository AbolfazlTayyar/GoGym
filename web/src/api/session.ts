import type { QueryClient } from "@tanstack/react-query";

export const signInPath = "/login";
export const nextParam = "next";

type Navigate = (to: string, options: { replace: boolean }) => void;

// A 401 means the cookie is gone or expired: drop everything cached and sign in again, then come back here.
// On the sign-in page itself a 401 is the expected answer, and there is nothing to clear.
export function goToSignIn(navigate: Navigate, queryClient: QueryClient, current: string) {
  if (current.startsWith(signInPath)) return;
  queryClient.clear();
  navigate(`${signInPath}?${nextParam}=${encodeURIComponent(current)}`, { replace: true });
}

// Only a path inside the app, so a crafted link can't send the coach to another site after signing in.
export function safeNext(next: string | null): string {
  return next && next.startsWith("/") && !next.startsWith("//") ? next : "/";
}
