import { describe, expect, it } from "vitest";
import { dayColumns, daySlot, letterColumns } from "./days";

describe("daySlot", () => {
  it.each([
    ["A", 0],
    ["E", 4],
    ["G", 6],
    ["day1", 0],
    ["day7", 6],
    ["H", -1],
    ["AB", -1],
  ])("%s → %i", (label, slot) => {
    expect(daySlot(label)).toBe(slot);
  });
});

describe("letterColumns", () => {
  it("leaves the fifth place empty for a four-day plan", () => {
    expect(letterColumns(["A", "B", "C", "D"], 5)).toEqual(["A", "B", "C", "D", null]);
  });

  it("keeps a deleted day's place empty", () => {
    expect(letterColumns(["A", "B", "D", "E"], 5)).toEqual(["A", "B", null, "D", "E"]);
  });

  it("places labels by slot, not by list position", () => {
    expect(letterColumns(["day2", "day1"], 5)).toEqual(["day1", "day2", null, null, null]);
  });
});

describe("dayColumns", () => {
  it("is five for plans of up to five days", () => {
    expect(dayColumns([["A", "B"], ["A", "B", "C", "D", "E"], []])).toBe(5);
  });

  it("widens for a seven-day plan", () => {
    expect(dayColumns([["A"], ["A", "B", "C", "D", "E", "F", "G"]])).toBe(7);
  });
});
