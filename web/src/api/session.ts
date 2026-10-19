import type { QueryClient } from "@tanstack/react-query";

export const signInPath = "/login";
export const nextParam = "next";

type Navigate = (to: string, options: { replace: boolean }) => void;

// A 401 means the cookie is gone or expired: drop everything cached and sign in again, then come back here.
export function goToSignIn(navigate: Navigate, queryClient: QueryClient, current: string) {
  queryClient.clear();
  if (current.startsWith(signInPath)) return;
  navigate(`${signInPath}?${nextParam}=${encodeURIComponent(current)}`, { replace: true });
}
