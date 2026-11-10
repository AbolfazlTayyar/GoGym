import { toPersianDigits } from "../lib/digits";
import { msgNotFound } from "./codes";

// The Persian for the API's "at least one of … is required"; MeasurementSheet recognises that answer by it.
export const msgNoneFilled = "دست‌کم یکی را پر کنید.";

// The API's field messages are English (internal/*/service.go); each known shape gets a Persian one.
const rules: [RegExp, (m: RegExpMatchArray) => string][] = [
  [/^is required$/, () => "لازم است."],
  [/^at least one movement is required$/, () => "دست‌کم یک حرکت لازم است."],
  [/^at least one .* is required$/, () => msgNoneFilled],
  [/^must have at most (\d+) items$/, (m) => `حداکثر ${toPersianDigits(m[1])} مورد.`],
  [/^must be at most (\d+) characters$/, (m) => `حداکثر ${toPersianDigits(m[1])} نویسه.`],
  [/^must be at least (\d+) characters$/, (m) => `دست‌کم ${toPersianDigits(m[1])} نویسه.`],
  [/^must be an Iranian mobile number/, () => "شماره موبایل درست نیست؛ مثلاً ۰۹۱۲۳۴۵۶۷۸۹."],
  [/^must be a date/, () => "تاریخ درست نیست."],
  [/^must be between \d{4}-\d\d-\d\d and \d{4}-\d\d-\d\d$/, () => "این تاریخ مجاز نیست."],
  [
    /^must be between ([\d.]+) and ([\d.]+)/,
    (m) => `باید بین ${toPersianDigits(m[1])} و ${toPersianDigits(m[2])} باشد.`,
  ],
  [/^must be an integer$/, () => "باید عدد صحیح باشد."],
  [/^not found$/, () => msgNotFound],
  [/^is already used/, () => "تکراری است."],
];

export function translateFieldMessage(message: string): string {
  for (const [pattern, toPersian] of rules) {
    const match = message.match(pattern);
    if (match) return toPersian(match);
  }
  return "مقدار درست نیست.";
}
