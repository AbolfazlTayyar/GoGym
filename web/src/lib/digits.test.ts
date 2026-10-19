import { describe, expect, it } from "vitest";
import { formatPhone, parseNumber, toLatinDigits, toPersianDigits, toPhone } from "./digits";

describe("toLatinDigits", () => {
  it.each([
    ["۰۹۱۲۳۴۵۶۷۸۹", "09123456789"],
    ["٠٩١٢٣٤٥٦٧٨٩", "09123456789"],
    ["۰9١2۳", "09123"],
    ["۶۸٫۵", "68.5"],
    ["نام ۱۲", "نام 12"],
  ])("%s → %s", (input, expected) => {
    expect(toLatinDigits(input)).toBe(expected);
  });
});

describe("toPersianDigits", () => {
  it("shows digits and the decimal separator in Persian", () => {
    expect(toPersianDigits(68.5)).toBe("۶۸٫۵");
    expect(toPersianDigits("1405")).toBe("۱۴۰۵");
  });
});

describe("parseNumber", () => {
  it.each([
    ["۶۸٫۵", 68.5],
    ["٦٨٫٥", 68.5],
    [" 164 ", 164],
    ["", null],
    ["  ", null],
    ["۶۸٫", null],
    ["abc", null],
    ["1.2.3", null],
  ])("%j → %j", (input, expected) => {
    expect(parseNumber(input)).toBe(expected);
  });
});

describe("toPhone", () => {
  it("drops grouping and converts digits", () => {
    expect(toPhone("۰۹۱۲ ۳۴۵-۶۷۸۹")).toBe("09123456789");
  });
});

describe("formatPhone", () => {
  it("groups a mobile number 4-3-4 in Persian digits", () => {
    expect(formatPhone("09123456789")).toBe("۰۹۱۲ ۳۴۵ ۶۷۸۹");
  });
});
