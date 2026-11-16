import { toPersianDigits } from "./digits";

// The coach's cue to collect the next payment and write a new plan.
export const msgSessionsDone = "تمام شد";

// ۷ از ۱۲
export function visitCountText(visits: number, sessions: number): string {
  return `${toPersianDigits(visits)} از ${toPersianDigits(sessions)}`;
}

export function sessionsUsed(visits: number, sessions: number): boolean {
  return visits >= sessions;
}
