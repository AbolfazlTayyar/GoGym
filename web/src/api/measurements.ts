import { api } from "./client";
import type { Measurement } from "./types";

// An athlete's check-ins, oldest first: the order a trend chart plots in.
export const listMeasurements = (athleteId: string) =>
  api.get<Measurement[]>(`/athletes/${athleteId}/measurements`);

// A check-in: the date, and only the measurements taken; PUT replaces them all, so one sent as null is cleared.
export type MeasurementInput = {
  date: string;
  weight: number | null;
  chest: number | null;
  waist: number | null;
  arm: number | null;
  thigh: number | null;
  hip: number | null;
};

export const createMeasurement = (athleteId: string, input: MeasurementInput) =>
  api.post<Measurement>(`/athletes/${athleteId}/measurements`, input);

export const updateMeasurement = (id: string, input: MeasurementInput) =>
  api.put<Measurement>(`/measurements/${id}`, input);

export const deleteMeasurement = (id: string) => api.delete(`/measurements/${id}`);
