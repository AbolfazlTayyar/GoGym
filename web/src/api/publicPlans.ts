import { api } from "./client";
import type { SharedPlan } from "./types";

// Needs no login: the token is the access, so this never answers 401.
export const getSharedPlan = (token: string) =>
  api.get<SharedPlan>(`/public/plans/${encodeURIComponent(token)}`);
