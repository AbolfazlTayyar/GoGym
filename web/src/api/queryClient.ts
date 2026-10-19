import { QueryClient } from "@tanstack/react-query";
import { ApiError } from "./client";
import { ErrorCode } from "./codes";

const maxNetworkRetries = 2;

// Only a request that never got an answer is worth retrying: a 4xx won't change on retry.
export function shouldRetry(failureCount: number, error: unknown): boolean {
  return (
    error instanceof ApiError &&
    error.code === ErrorCode.network &&
    failureCount < maxNetworkRetries
  );
}

// networkMode "always": offline, a request fails at once with the network error instead of waiting,
// paused, for the connection; a failed save then shows and keeps the draft rather than hanging.
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: { refetchOnWindowFocus: false, retry: shouldRetry, networkMode: "always" },
    mutations: { retry: false, networkMode: "always" },
  },
});
