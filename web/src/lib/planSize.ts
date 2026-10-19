import type { PlanSummary } from "../api/types";
import { toPersianDigits } from "./digits";

// "۵ روز · ۴۳ حرکت", or "خالی" for a plan with no days yet.
export function planSizeText(plan: Pick<PlanSummary, "day_labels" | "block_count">): string {
  if (plan.day_labels.length === 0) return "خالی";
  return `${toPersianDigits(plan.day_labels.length)} روز · ${toPersianDigits(plan.block_count)} حرکت`;
}
