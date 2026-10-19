import { api } from "./client";
import type { Measurement } from "./types";

// An athlete's check-ins, oldest first: the order a trend chart plots in.
export const listMeasurements = (athleteId: string) =>
  api.get<Measurement[]>(`/athletes/${athleteId}/measurements`);
