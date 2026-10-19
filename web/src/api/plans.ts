import { api } from "./client";
import type { Day, PlanDetail, PlanSummary } from "./types";

// Every athlete's current plan in one request; the home rows match them by athlete_id.
export const getCurrentPlans = () => api.get<PlanSummary[]>("/plans/current");

export const getPlan = (id: string) => api.get<PlanDetail>(`/plans/${id}`);

export const listAthletePlans = (athleteId: string) =>
  api.get<PlanSummary[]>(`/athletes/${athleteId}/plans`);

export type CreatePlanInput = {
  athlete_id: string;
  title: string;
  start_date: string;
  // A copy of this plan's days; omitted for an empty plan.
  source_plan_id?: string;
};

export const createPlan = (input: CreatePlanInput) => api.post<PlanDetail>("/plans", input);

export const addDay = (planId: string, label: string, orderIndex: number) =>
  api.post<Day>(`/plans/${planId}/days`, { label, order_index: orderIndex });

// The body of PUT /days/:id/blocks: the whole day, in order; positions are the order.
export type SaveBlock = {
  sets: number;
  rest_seconds: number | null;
  notes: string | null;
  movements: {
    movement_id: string;
    reps: number | null;
    duration_seconds: number | null;
    load: string | null;
  }[];
};

export const saveDay = (dayId: string, blocks: SaveBlock[]) =>
  api.put<Day>(`/days/${dayId}/blocks`, blocks);

export const deleteDay = (dayId: string) => api.delete(`/days/${dayId}`);

export type PlanUpdate = { title: string; start_date: string; note: string | null };

export const updatePlan = (id: string, input: PlanUpdate) =>
  api.put<PlanDetail>(`/plans/${id}`, input);

export const deletePlan = (id: string) => api.delete(`/plans/${id}`);
