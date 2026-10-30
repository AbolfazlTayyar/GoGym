import { jalaaliMonthLength, toGregorian, toJalaali } from "jalaali-js";

export const jalaliMonthNames = [
  "فروردین",
  "اردیبهشت",
  "خرداد",
  "تیر",
  "مرداد",
  "شهریور",
  "مهر",
  "آبان",
  "آذر",
  "دی",
  "بهمن",
  "اسفند",
];

export type JalaliDate = { jy: number; jm: number; jd: number };

const pad = (n: number) => String(n).padStart(2, "0");

// An ISO date is a calendar day, not an instant: it's read and shown in UTC so no time zone can shift it.
function isoToUTCDate(iso: string): Date {
  const [y, m, d] = iso.split("-").map(Number);
  return new Date(Date.UTC(y, m - 1, d));
}

const longFormat = new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
  day: "numeric",
  month: "long",
  year: "numeric",
  timeZone: "UTC",
});

// 2026-09-23 → ۱ مهر ۱۴۰۵
export function formatJalali(iso: string): string {
  return longFormat.format(isoToUTCDate(iso));
}

const shortFormat = new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
  day: "numeric",
  month: "long",
  timeZone: "UTC",
});

// 2026-09-23 → ۱ مهر, for chart labels.
export function formatJalaliShort(iso: string): string {
  return shortFormat.format(isoToUTCDate(iso));
}

export function isoToJalali(iso: string): JalaliDate {
  const [y, m, d] = iso.split("-").map(Number);
  return toJalaali(y, m, d);
}

export function jalaliToISO({ jy, jm, jd }: JalaliDate): string {
  const { gy, gm, gd } = toGregorian(jy, jm, jd);
  return `${gy}-${pad(gm)}-${pad(gd)}`;
}

export function jalaliMonthLength(jy: number, jm: number): number {
  return jalaaliMonthLength(jy, jm);
}

// Today on the phone's clock, not UTC: a check-in logged at 01:00 in Tehran is today's, not yesterday's.
export function todayISO(now: Date = new Date()): string {
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`;
}

export function jalaliMonthName(iso: string): string {
  return jalaliMonthNames[isoToJalali(iso).jm - 1];
}

// The coach's titles name the month: "برنامه آبان".
export function planTitleFor(iso: string): string {
  return `برنامه ${jalaliMonthName(iso)}`;
}
