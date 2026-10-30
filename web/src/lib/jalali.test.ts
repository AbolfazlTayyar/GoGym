import { afterEach, describe, expect, it, vi } from "vitest";
import {
  formatJalali,
  formatJalaliShort,
  isoToJalali,
  jalaliMonthLength,
  jalaliToISO,
  jalaliWeekdayIndex,
  jalaliWeekdayNames,
  planTitleFor,
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
    expect(formatJalaliShort("2026-05-24")).toBe("۳ خرداد");
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

describe("jalaliWeekdayIndex", () => {
  it("counts from Saturday", () => {
    expect(jalaliWeekdayNames[jalaliWeekdayIndex("2026-10-24")]).toBe("شنبه");
    expect(jalaliWeekdayNames[jalaliWeekdayIndex("2026-09-23")]).toBe("چهارشنبه");
    expect(jalaliWeekdayNames[jalaliWeekdayIndex("2026-10-30")]).toBe("جمعه");
  });
});

describe("planTitleFor", () => {
  it("names the Jalali month the date falls in", () => {
    expect(planTitleFor("2026-10-30")).toBe("برنامه آبان");
    expect(planTitleFor("2026-03-21")).toBe("برنامه فروردین");
  });
});
