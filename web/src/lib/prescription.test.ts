import { describe, expect, it } from "vitest";
import { prescriptionText } from "./prescription";

const part = (p: Partial<{ reps: number; duration_seconds: number; load: string }>) => ({
  reps: p.reps ?? null,
  duration_seconds: p.duration_seconds ?? null,
  load: p.load ?? null,
});

describe("prescriptionText", () => {
  it("writes reps × sets", () => {
    expect(prescriptionText({ sets: 3, movements: [part({ reps: 15 })] })).toBe("۱۵ تکرار × ۳ ست");
  });

  it("writes seconds", () => {
    expect(prescriptionText({ sets: 3, movements: [part({ duration_seconds: 45 })] })).toBe(
      "۴۵ ثانیه × ۳ ست",
    );
  });

  it("writes whole minutes as minutes", () => {
    expect(prescriptionText({ sets: 2, movements: [part({ duration_seconds: 120 })] })).toBe(
      "۲ دقیقه × ۲ ست",
    );
    expect(prescriptionText({ sets: 1, movements: [part({ duration_seconds: 90 })] })).toBe(
      "۹۰ ثانیه × ۱ ست",
    );
  });

  it("puts the load after its movement", () => {
    expect(prescriptionText({ sets: 4, movements: [part({ reps: 8, load: "۲۰ کیلو" })] })).toBe(
      "۸ تکرار (۲۰ کیلو) × ۴ ست",
    );
  });

  it("joins a three-movement block with +", () => {
    expect(
      prescriptionText({
        sets: 3,
        movements: [
          part({ reps: 12 }),
          part({ reps: 12, load: "بدون وزنه" }),
          part({ duration_seconds: 30 }),
        ],
      }),
    ).toBe("۱۲ تکرار + ۱۲ تکرار (بدون وزنه) + ۳۰ ثانیه × ۳ ست");
  });

  it("falls back to sets alone when no amount is set", () => {
    expect(prescriptionText({ sets: 3, movements: [part({})] })).toBe("۳ ست");
  });
});
