const persianDigits = "۰۱۲۳۴۵۶۷۸۹";
const arabicDigits = "٠١٢٣٤٥٦٧٨٩";

// A Persian phone keyboard types ۱۲ and ٫, which the API rejects; every number and phone a coach types goes through this.
export function toLatinDigits(text: string): string {
  return text
    .replace(/[۰-۹]/g, (d) => String(persianDigits.indexOf(d)))
    .replace(/[٠-٩]/g, (d) => String(arabicDigits.indexOf(d)))
    .replace(/٫/g, ".");
}

// For display: 12.5 → ۱۲٫۵.
export function toPersianDigits(value: string | number): string {
  return String(value)
    .replace(/\d/g, (d) => persianDigits[Number(d)])
    .replace(/\./g, "٫");
}

// null for an empty field or anything that isn't a plain non-negative number.
export function parseNumber(text: string): number | null {
  const latin = toLatinDigits(text.trim());
  return /^\d+(\.\d+)?$/.test(latin) ? Number(latin) : null;
}
