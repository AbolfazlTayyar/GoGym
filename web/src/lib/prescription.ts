import { toPersianDigits } from "./digits";

// The parts of a block the text needs; the day view passes API blocks, the builder its draft lines.
type Part = { reps: number | null; duration_seconds: number | null; load: string | null };
type Prescribed = { sets: number; movements: Part[] };

const secondsPerMinute = 60;

// One movement's amount in the coach's wording: "۱۵ تکرار", "۴۵ ثانیه", whole minutes as "۲ دقیقه".
export function amountText(part: Part): string {
  let amount = "";
  if (part.duration_seconds !== null) {
    const s = part.duration_seconds;
    amount =
      s % secondsPerMinute === 0
        ? `${toPersianDigits(s / secondsPerMinute)} دقیقه`
        : `${toPersianDigits(s)} ثانیه`;
  } else if (part.reps !== null) {
    amount = `${toPersianDigits(part.reps)} تکرار`;
  }
  // The load follows its own movement, so "10 with weight + 10 without" in one block reads apart.
  return part.load ? `${amount} (${part.load})`.trim() : amount;
}

// A block's line under its movements, as her plans write it: "۱۵ تکرار + ۴۵ ثانیه × ۳ ست".
export function prescriptionText(block: Prescribed): string {
  const amounts = block.movements.map(amountText).filter(Boolean).join(" + ");
  const sets = `${toPersianDigits(block.sets)} ست`;
  return amounts ? `${amounts} × ${sets}` : sets;
}
