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
