import { api } from "./client";
import type { Movement } from "./types";

// The coach's own movements and the universal ones together, ordered by name.
export function listMovements(q = "") {
  const params = q ? `?${new URLSearchParams({ q })}` : "";
  return api.get<Movement[]>(`/movements${params}`);
}

export type MovementInput = {
  name: string;
  category: string | null;
  description: string | null;
};

export const createMovement = (input: MovementInput) => api.post<Movement>("/movements", input);

export const updateMovement = (id: string, input: MovementInput) =>
  api.put<Movement>(`/movements/${id}`, input);

export const deleteMovement = (id: string) => api.delete(`/movements/${id}`);
