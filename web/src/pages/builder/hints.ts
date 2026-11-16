import type { Day, MovementHistory } from "../../api/types";
import { prescriptionText } from "../../lib/prescription";
import type { Draft } from "./draft";

// The search chips' second line, by movement id: whether this plan already has the movement, and what the
// athlete was last given of it in another plan. One pass over data already loaded; no request per chip.
export function movementHints(
  history: MovementHistory[],
  otherDays: Day[],
  draft: Draft,
): Map<string, string> {
  // This plan is the saved other days plus the day being edited as it stands, unsaved lines included.
  const inPlan = new Set<string>();
  for (const d of otherDays)
    for (const b of d.blocks) for (const m of b.movements) inPlan.add(m.movement_id);
  for (const b of draft) for (const m of b.movements) inPlan.add(m.movement_id);

  const hints = new Map<string, string>();
  for (const id of inPlan) hints.set(id, "در این برنامه");
  for (const h of history) {
    const last = prescriptionText({ sets: h.sets, movements: [h] });
    const here = hints.get(h.movement_id);
    hints.set(h.movement_id, here ? `${here} · ${last}` : last);
  }
  return hints;
}
