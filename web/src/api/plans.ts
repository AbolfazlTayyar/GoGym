import { api } from "./client";
import type { PlanSummary } from "./types";

// Every athlete's current plan in one request; the home rows match them by athlete_id.
export const getCurrentPlans = () => api.get<PlanSummary[]>("/plans/current");
