// Arabic keyboards type ي, ى and ك where Persian has ی and ک.
const persianLetters: Record<string, string> = {
  ي: "ی",
  ى: "ی",
  ك: "ک",
};

// The same rules as the API's, so the chip the coach sees is the category as saved, and a rename can
// tell up front that it will merge into a category already in use.
export function normalizeCategory(value: string): string {
  return value
    .replace(/[يىك]/g, (c) => persianLetters[c])
    .split(/\s+/)
    .filter(Boolean)
    .join(" ")
    .replace(/^[\s‌]+|[\s‌]+$/g, "");
}
