import type { SaveBlock } from "../../api/plans";
import type { Day } from "../../api/types";

// The day being edited, held in the browser until ذخیره sends it whole (PUT /days/:id/blocks).
// Every edit is a pure function of the draft; keys are given by the caller, so the reducer stays pure.

export type DraftMovement = {
  key: string;
  movement_id: string;
  name: string;
  reps: number | null;
  duration_seconds: number | null;
  load: string | null;
};

export type DraftBlock = {
  key: string;
  sets: number;
  rest_seconds: number | null;
  notes: string | null;
  movements: DraftMovement[];
};

export type Draft = DraftBlock[];

export type Mode = "reps" | "seconds";

// The coach's common prescriptions (docs/design/src/data.js PRESETS).
export const presets = [
  { reps: 12, sets: 3 },
  { reps: 15, sets: 3 },
  { reps: 20, sets: 3 },
  { reps: 15, sets: 4 },
  { seconds: 30, sets: 3 },
] as const;
export type Preset = (typeof presets)[number];

// What a line starts as when there is no line before it to copy.
const firstLine = { sets: 3, reps: 12 };
const defaultSeconds = 30;

const bounds = {
  sets: { min: 1, max: 50, step: 1 },
  reps: { min: 1, max: 1000, step: 1 },
  seconds: { min: 5, max: 4 * 60 * 60, step: 5 },
};

export type Action =
  | { type: "load"; day: Day }
  | { type: "add"; blockKey: string; movementKey: string; movement: { id: string; name: string } }
  | { type: "join"; blockKey: string; movementKey: string; movement: { id: string; name: string } }
  | { type: "mode"; blockKey: string; movementKey: string; mode: Mode }
  | { type: "step"; blockKey: string; movementKey: string; delta: 1 | -1 }
  | { type: "sets"; blockKey: string; delta: 1 | -1 }
  | { type: "preset"; blockKey: string; preset: Preset }
  | { type: "move"; blockKey: string; delta: 1 | -1 }
  | { type: "remove"; blockKey: string }
  | { type: "removeMovement"; blockKey: string; movementKey: string }
  | { type: "split"; blockKey: string; keys: string[] };

export function fromDay(day: Day): Draft {
  return day.blocks.map((b) => ({
    key: b.id,
    sets: b.sets,
    rest_seconds: b.rest_seconds,
    notes: b.notes,
    movements: b.movements.map((m) => ({
      key: m.id,
      movement_id: m.movement_id,
      name: m.name,
      reps: m.reps,
      duration_seconds: m.duration_seconds,
      load: m.load,
    })),
  }));
}

export function toBody(draft: Draft): SaveBlock[] {
  return draft.map((b) => ({
    sets: b.sets,
    rest_seconds: b.rest_seconds,
    notes: b.notes,
    movements: b.movements.map((m) => ({
      movement_id: m.movement_id,
      reps: m.reps,
      duration_seconds: m.duration_seconds,
      load: m.load,
    })),
  }));
}

export function modeOf(m: DraftMovement): Mode {
  return m.duration_seconds !== null ? "seconds" : "reps";
}

export function draftReducer(draft: Draft, action: Action): Draft {
  switch (action.type) {
    case "load":
      return fromDay(action.day);
    case "add":
      return addLine(draft, action.blockKey, action.movementKey, action.movement);
    case "join":
      return joinLine(draft, action.blockKey, action.movementKey, action.movement);
    case "mode":
      return editMovement(draft, action.blockKey, action.movementKey, (m) =>
        setMode(m, action.mode),
      );
    case "step":
      return editMovement(draft, action.blockKey, action.movementKey, (m) => step(m, action.delta));
    case "sets":
      return editBlock(draft, action.blockKey, (b) => ({
        ...b,
        sets: clamp(b.sets + action.delta * bounds.sets.step, bounds.sets),
      }));
    case "preset":
      return editBlock(draft, action.blockKey, (b) => applyPreset(b, action.preset));
    case "move":
      return move(draft, action.blockKey, action.delta);
    case "remove":
      return draft.filter((b) => b.key !== action.blockKey);
    case "removeMovement":
      return removeMovement(draft, action.blockKey, action.movementKey);
    case "split":
      return split(draft, action.blockKey, action.keys);
  }
}

// A new line copies the previous line's sets and its first movement's amount, the way her days repeat.
function addLine(
  draft: Draft,
  blockKey: string,
  movementKey: string,
  movement: { id: string; name: string },
): Draft {
  const previous = draft.at(-1);
  const amount = previous?.movements[0];
  return [
    ...draft,
    {
      key: blockKey,
      sets: previous?.sets ?? firstLine.sets,
      rest_seconds: null,
      notes: null,
      movements: [
        {
          key: movementKey,
          movement_id: movement.id,
          name: movement.name,
          reps: amount ? amount.reps : firstLine.reps,
          duration_seconds: amount ? amount.duration_seconds : null,
          load: null,
        },
      ],
    },
  ];
}

// Joining makes a superset: the movement becomes part of the line, with the amount of the one before it.
function joinLine(
  draft: Draft,
  blockKey: string,
  movementKey: string,
  movement: { id: string; name: string },
): Draft {
  return editBlock(draft, blockKey, (b) => {
    const last = b.movements.at(-1);
    return {
      ...b,
      movements: [
        ...b.movements,
        {
          key: movementKey,
          movement_id: movement.id,
          name: movement.name,
          reps: last ? last.reps : firstLine.reps,
          duration_seconds: last ? last.duration_seconds : null,
          load: null,
        },
      ],
    };
  });
}

function setMode(m: DraftMovement, mode: Mode): DraftMovement {
  if (modeOf(m) === mode) return m;
  return mode === "seconds"
    ? { ...m, reps: null, duration_seconds: defaultSeconds }
    : { ...m, reps: firstLine.reps, duration_seconds: null };
}

function step(m: DraftMovement, delta: 1 | -1): DraftMovement {
  if (modeOf(m) === "seconds") {
    const b = bounds.seconds;
    // Off-step values (a copied 45) snap to the step grid on the first tap.
    const snapped = Math.round(m.duration_seconds! / b.step) * b.step;
    return { ...m, duration_seconds: clamp(snapped + delta * b.step, b) };
  }
  return { ...m, reps: clamp((m.reps ?? firstLine.reps) + delta * bounds.reps.step, bounds.reps) };
}

// A preset sets the line's sets and gives every movement in it the same amount.
function applyPreset(b: DraftBlock, preset: Preset): DraftBlock {
  return {
    ...b,
    sets: preset.sets,
    movements: b.movements.map((m) =>
      "seconds" in preset
        ? { ...m, reps: null, duration_seconds: preset.seconds }
        : { ...m, reps: preset.reps, duration_seconds: null },
    ),
  };
}

export function presetMatches(b: DraftBlock, preset: Preset): boolean {
  return (
    b.sets === preset.sets &&
    b.movements.every((m) =>
      "seconds" in preset
        ? m.duration_seconds === preset.seconds
        : m.reps === preset.reps && m.duration_seconds === null,
    )
  );
}

// A superset is one block, so moving past it swaps whole lines and never splits it.
function move(draft: Draft, blockKey: string, delta: 1 | -1): Draft {
  const from = draft.findIndex((b) => b.key === blockKey);
  const to = from + delta;
  if (from < 0 || to < 0 || to >= draft.length) return draft;
  const next = [...draft];
  [next[from], next[to]] = [next[to], next[from]];
  return next;
}

// Taking the last movement out of a line takes the line too.
function removeMovement(draft: Draft, blockKey: string, movementKey: string): Draft {
  return draft.flatMap((b) => {
    if (b.key !== blockKey) return [b];
    const movements = b.movements.filter((m) => m.key !== movementKey);
    return movements.length ? [{ ...b, movements }] : [];
  });
}

// Breaking a superset gives each movement its own line, in place, with the superset's sets.
// keys name the new lines; the first keeps the block's own key.
function split(draft: Draft, blockKey: string, keys: string[]): Draft {
  return draft.flatMap((b) => {
    if (b.key !== blockKey) return [b];
    return b.movements.map((m, i) => ({
      ...b,
      key: i === 0 ? b.key : keys[i - 1],
      movements: [m],
    }));
  });
}

function editBlock(draft: Draft, blockKey: string, edit: (b: DraftBlock) => DraftBlock): Draft {
  return draft.map((b) => (b.key === blockKey ? edit(b) : b));
}

function editMovement(
  draft: Draft,
  blockKey: string,
  movementKey: string,
  edit: (m: DraftMovement) => DraftMovement,
): Draft {
  return editBlock(draft, blockKey, (b) => ({
    ...b,
    movements: b.movements.map((m) => (m.key === movementKey ? edit(m) : m)),
  }));
}

function clamp(value: number, { min, max }: { min: number; max: number }): number {
  return Math.min(max, Math.max(min, value));
}

let lastKey = 0;
// Keys for new lines and movements; the reducer never makes them, so it stays pure.
export function newKey(): string {
  lastKey += 1;
  return `new-${lastKey}`;
}
