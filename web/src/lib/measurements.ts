import type { Measurement } from "../api/types";
import { toPersianDigits } from "./digits";

export type Metric = "weight" | "chest" | "waist" | "arm" | "thigh" | "hip";

// In the order the chips and the sheet show them; weight first.
export const metrics: { key: Metric; label: string; unit: string }[] = [
  { key: "weight", label: "وزن", unit: "کیلو" },
  { key: "chest", label: "سینه", unit: "سانت" },
  { key: "waist", label: "کمر", unit: "سانت" },
  { key: "arm", label: "بازو", unit: "سانت" },
  { key: "thigh", label: "ران", unit: "سانت" },
  { key: "hip", label: "باسن", unit: "سانت" },
];

// 69.6 → ۶۹٫۶, 92 → ۹۲: at most one decimal, as she measures.
export function formatValue(value: number): string {
  return toPersianDigits(Number(value.toFixed(1)));
}

// The check-ins that have a value for metric, oldest first; one without it is skipped, not drawn as 0.
export function series(
  measurements: Measurement[],
  metric: Metric,
): { date: string; value: number }[] {
  return measurements.flatMap((m) =>
    m[metric] === null ? [] : [{ date: m.date, value: m[metric] }],
  );
}

// The value a field shows under it while the coach types: the latest check-in before this one that has
// it. measurements are oldest first; editing passes the check-in being edited, so only earlier ones count.
export function lastValue(
  measurements: Measurement[],
  metric: Metric,
  editing?: Measurement,
): number | null {
  const end = editing ? measurements.findIndex((m) => m.id === editing.id) : measurements.length;
  for (let i = (end < 0 ? measurements.length : end) - 1; i >= 0; i--) {
    const value = measurements[i][metric];
    if (value !== null) return value;
  }
  return null;
}
