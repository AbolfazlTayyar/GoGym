import { afterEach, describe, expect, it, vi } from "vitest";
import {
  firstOfNextJalaliMonth,
  formatJalali,
  isoToJalali,
  jalaliMonthLength,
  jalaliMonthName,
  jalaliToISO,
  todayISO,
} from "./jalali";

describe("Jalali ↔ ISO", () => {
  it("converts Nowruz 1405", () => {
    expect(jalaliToISO({ jy: 1405, jm: 1, jd: 1 })).toBe("2026-03-21");
    expect(isoToJalali("2026-03-21")).toEqual({ jy: 1405, jm: 1, jd: 1 });
  });

  it("knows 1403 is a leap year", () => {
    expect(jalaliMonthLength(1403, 12)).toBe(30);
    expect(jalaliToISO({ jy: 1403, jm: 12, jd: 30 })).toBe("2025-03-20");
    expect(jalaliMonthLength(1404, 12)).toBe(29);
  });
});

describe("formatJalali", () => {
  it("shows the day, month name and year in Persian digits", () => {
    expect(formatJalali("2026-09-23")).toBe("۱ مهر ۱۴۰۵");
    expect(formatJalali("2025-03-20")).toBe("۳۰ اسفند ۱۴۰۳");
  });
});

describe("todayISO", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("uses the phone's local date just after midnight in Tehran", () => {
    vi.stubEnv("TZ", "Asia/Tehran");
    // 01:00 on 19 Oct in Tehran is still 18 Oct in UTC.
    expect(todayISO(new Date("2026-10-18T21:30:00Z"))).toBe("2026-10-19");
  });
});

describe("firstOfNextJalaliMonth", () => {
  it("moves to the 1st of the next month", () => {
    expect(firstOfNextJalaliMonth("2026-10-19")).toBe("2026-10-23"); // 27 Mehr → 1 Aban
  });

  it("crosses Esfand into Farvardin of the next year", () => {
    expect(firstOfNextJalaliMonth("2026-03-01")).toBe("2026-03-21"); // 10 Esfand 1404 → 1 Farvardin 1405
    expect(jalaliMonthName(firstOfNextJalaliMonth("2026-03-01"))).toBe("فروردین");
  });
});
