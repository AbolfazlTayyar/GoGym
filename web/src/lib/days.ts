// A plan's days are labelled A-G or day1-day7, one per slot (backend: plan.dayLabels).
const letters = "ABCDEFG";

// The home rows' letters keep their columns: a four-day plan leaves E's place empty.
export const minDayColumns = 5;

// daySlot is the column a day's letter sits in: A or day1 is 0. -1 for a label outside both schemes.
export function daySlot(label: string): number {
  const letter = letters.indexOf(label);
  if (label.length === 1 && letter >= 0) return letter;
  const numbered = /^day([1-7])$/.exec(label);
  return numbered ? Number(numbered[1]) - 1 : -1;
}

// letterColumns places each label in its slot, leaving null where the plan has no day.
export function letterColumns(labels: string[], columns: number): (string | null)[] {
  const placed: (string | null)[] = Array.from({ length: columns }, () => null);
  for (const label of labels) {
    const slot = daySlot(label);
    if (slot >= 0 && slot < columns) placed[slot] = label;
  }
  return placed;
}

// dayColumns is five, or more when any listed plan uses a sixth or seventh slot, so every row lines up.
export function dayColumns(labelLists: string[][]): number {
  let columns = minDayColumns;
  for (const labels of labelLists) {
    for (const label of labels) columns = Math.max(columns, daySlot(label) + 1);
  }
  return columns;
}
