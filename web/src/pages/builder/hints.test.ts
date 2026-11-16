import { describe, expect, it } from "vitest";
import type { Day, MovementHistory } from "../../api/types";
import type { Draft } from "./draft";
import { movementHints } from "./hints";

const last = (movement_id: string, rest: Partial<MovementHistory>): MovementHistory => ({
  movement_id,
  plan_id: "p0",
  plan_title: "مهر",
  start_date: "2026-10-01",
  sets: 3,
  reps: null,
  duration_seconds: null,
  load_kg: null,
  ...rest,
});

const savedDay = (movement_id: string): Day => ({
  id: "d2",
  label: "B",
  order_index: 1,
  blocks: [
    {
      id: "b",
      order_index: 0,
      sets: 3,
      rest_seconds: null,
      notes: null,
      movements: [
        {
          id: "bm",
          movement_id,
          name: "",
          category: null,
          reps: 10,
          duration_seconds: null,
          load_kg: null,
          order_in_block: 0,
        },
      ],
    },
  ],
});

const draftWith = (movement_id: string): Draft => [
  {
    key: "k",
    sets: 3,
    rest_seconds: null,
    notes: null,
    movements: [
      { key: "mk", movement_id, name: "", reps: 12, duration_seconds: null, load_kg: null },
    ],
  },
];

describe("movementHints", () => {
  it("writes the last prescription from another plan in the day view's words", () => {
    const hints = movementHints(
      [last("squat", { reps: 10, load_kg: 40 }), last("plank", { duration_seconds: 45 })],
      [],
      [],
    );

    expect(hints.get("squat")).toBe("۱۰ تکرار (۴۰ کیلو) × ۳ ست");
    expect(hints.get("plank")).toBe("۴۵ ثانیه × ۳ ست");
  });

  it("marks a movement on another day of this plan, or in the unsaved draft", () => {
    const hints = movementHints([], [savedDay("lunge")], draftWith("row"));

    expect(hints.get("lunge")).toBe("در این برنامه");
    expect(hints.get("row")).toBe("در این برنامه");
    expect(hints.has("squat")).toBe(false);
  });

  it("puts this plan first when the movement is also in an earlier one", () => {
    const hints = movementHints([last("row", { reps: 8 })], [], draftWith("row"));

    expect(hints.get("row")).toBe("در این برنامه · ۸ تکرار × ۳ ست");
  });
});
