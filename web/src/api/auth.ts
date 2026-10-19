import { useQuery } from "@tanstack/react-query";
import { api, loginPath } from "./client";
import { queryKeys } from "./keys";
import type { Coach } from "./types";

// The API sets and clears the httpOnly auth cookie; the app never sees the token (docs/adr/0031).
export const login = (phone: string, password: string) =>
  api.post<null>(loginPath, { phone, password });

export const logout = () => api.post<null>("/auth/logout");

export const getMe = () => api.get<Coach>("/coaches/me");

// Whether the coach is signed in: the cookie is invisible to scripts, so only the API can say.
export function useCoach() {
  return useQuery({ queryKey: queryKeys.coach(), queryFn: getMe, staleTime: Infinity });
}
