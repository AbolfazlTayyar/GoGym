import { describe, expect, it } from "vitest";
import { normalizeCategory } from "./category";

describe("normalizeCategory", () => {
  it.each([
    ["کششي كشى", "کششی کشی"],
    ["  گرم   کردن ", "گرم کردن"],
    ["‌قدرتی ‌", "قدرتی"],
    ["بالا‌تنه", "بالا‌تنه"],
    [" ‌ ", ""],
  ])("%j becomes %j", (input, want) => {
    expect(normalizeCategory(input)).toBe(want);
  });
});
