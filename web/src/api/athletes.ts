import { api } from "./client";
import type { Athlete, ListMeta } from "./types";

// The coach has 20-30 athletes; the API's maximum page (docs/adr/0014) holds them all in one request.
export const allAthletesLimit = 100;

type ListParams = { type: "private" | "public" | "all"; q?: string };

export function listAthletes({ type, q }: ListParams) {
  const params = new URLSearchParams({ athlete_type: type, limit: String(allAthletesLimit) });
  if (q) params.set("q", q);
  return api.getWithMeta<Athlete[], ListMeta>(`/athletes?${params}`);
}
