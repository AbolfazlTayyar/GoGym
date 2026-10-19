import { describe, expect, it } from "vitest";
import type { Measurement } from "../api/types";
import { formatValue, lastValue, series } from "./measurements";

const m = (id: string, date: string, values: Partial<Measurement>): Measurement => ({
  id,
  athlete_id: "a1",
  date,
  weight: null,
  chest: null,
  waist: null,
  arm: null,
  thigh: null,
  hip: null,
  created_at: "",
  updated_at: "",
  ...values,
});

// Oldest first, as the API returns them.
const history = [
  m("1", "2026-05-24", { weight: 73.4, waist: 82 }),
  m("2", "2026-06-26", { weight: 72 }),
  m("3", "2026-07-24", { waist: 79.5 }),
];

describe("lastValue", () => {
  it("is the latest check-in that has the field", () => {
    expect(lastValue(history, "weight")).toBe(72);
    expect(lastValue(history, "waist")).toBe(79.5);
    expect(lastValue(history, "hip")).toBeNull();
  });

  it("only looks before the check-in being edited", () => {
    expect(lastValue(history, "waist", history[2])).toBe(82);
    expect(lastValue(history, "weight", history[0])).toBeNull();
  });
});

describe("series", () => {
  it("skips check-ins without the value instead of drawing 0", () => {
    expect(series(history, "waist")).toEqual([
      { date: "2026-05-24", value: 82 },
      { date: "2026-07-24", value: 79.5 },
    ]);
  });
});

describe("formatValue", () => {
  it("shows at most one decimal in Persian digits", () => {
    expect(formatValue(69.6)).toBe("۶۹٫۶");
    expect(formatValue(92)).toBe("۹۲");
    expect(formatValue(3.8000000000000043)).toBe("۳٫۸");
  });
});
