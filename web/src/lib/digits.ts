const persianDigits = "۰۱۲۳۴۵۶۷۸۹";

// For display: 12.5 → ۱۲٫۵.
export function toPersianDigits(value: string | number): string {
  return String(value)
    .replace(/\d/g, (d) => persianDigits[Number(d)])
    .replace(/\./g, "٫");
}
