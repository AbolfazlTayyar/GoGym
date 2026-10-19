import { api } from "./client";
import type { Athlete, AthleteType, ExperienceLevel, ListMeta } from "./types";

// The coach has 20-30 athletes; the API's maximum page (docs/adr/0014) holds them all in one request.
export const allAthletesLimit = 100;

type ListParams = { type: "private" | "public" | "all"; q?: string };

export function listAthletes({ type, q }: ListParams) {
  const params = new URLSearchParams({ athlete_type: type, limit: String(allAthletesLimit) });
  if (q) params.set("q", q);
  return api.getWithMeta<Athlete[], ListMeta>(`/athletes?${params}`);
}

// Every editable field: PUT replaces them all, so an optional field sent as null is cleared.
export type AthleteInput = {
  first_name: string;
  last_name: string;
  phone: string;
  athlete_type: AthleteType;
  experience_level: ExperienceLevel | null;
  goal: string | null;
  height: number | null;
  injuries: string | null;
};

export const getAthlete = (id: string) => api.get<Athlete>(`/athletes/${id}`);
export const createAthlete = (input: AthleteInput) => api.post<Athlete>("/athletes", input);
export const updateAthlete = (id: string, input: AthleteInput) =>
  api.put<Athlete>(`/athletes/${id}`, input);
export const deleteAthlete = (id: string) => api.delete(`/athletes/${id}`);
