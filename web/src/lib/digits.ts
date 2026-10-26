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

// A phone as the API wants it: Latin digits, without the spaces or dashes a coach may type to group it.
export function toPhone(text: string): string {
  return toLatinDigits(text).replace(/[\s-]/g, "");
}

// A mobile number is 09 plus nine digits.
const phoneLength = 11;

// For a phone field as it's typed: keeps only digits, in whichever script the keyboard types, and stops at a full number.
export function limitPhoneInput(text: string): string {
  return text.replace(/[^0-9۰-۹٠-٩]/g, "").slice(0, phoneLength);
}

// 09123456789 →۰۹۱۲ ۳۴۵ ۶۷۸۹, grouped the way the boards show a mobile number.
export function formatPhone(phone: string): string {
  const m = /^(\d{4})(\d{3})(\d{4})$/.exec(phone);
  return toPersianDigits(m ? `${m[1]} ${m[2]} ${m[3]}` : phone);
}
