import { describe, expect, it } from "vitest";
import type { Day } from "../../api/types";
import { draftReducer, fromDay, presetMatches, presets, stepKg, toBody, type Draft } from "./draft";

const lunge = { id: "m-lunge", name: "لانژ" };
const bridge = { id: "m-bridge", name: "پل باسن" };
const squat = { id: "m-squat", name: "اسکوات" };
const plank = { id: "m-plank", name: "پلانک" };

const day: Day = {
  id: "d1",
  label: "C",
  order_index: 2,
  blocks: [
    {
      id: "b1",
      order_index: 0,
      sets: 3,
      rest_seconds: null,
      notes: null,
      movements: [
        {
          id: "bm1",
          movement_id: lunge.id,
          name: lunge.name,
          category: null,
          reps: 12,
          duration_seconds: null,
          load_kg: null,
          order_in_block: 0,
        },
      ],
    },
    {
      id: "b2",
      order_index: 1,
      sets: 4,
      rest_seconds: 60,
      notes: "کند",
      movements: [
        {
          id: "bm2",
          movement_id: bridge.id,
          name: bridge.name,
          category: null,
          reps: null,
          duration_seconds: 45,
          load_kg: 10,
          order_in_block: 0,
        },
      ],
    },
  ],
};

const run = (draft: Draft, ...actions: Parameters<typeof draftReducer>[1][]) =>
  actions.reduce(draftReducer, draft);

// A short picture of a draft: "3×[لانژ 12r]" per line, "+" between a superset's movements.
const lines = (draft: Draft) =>
  draft.map(
    (b) =>
      `${b.sets}×[${b.movements
        .map((m) =>
          m.duration_seconds !== null ? `${m.name} ${m.duration_seconds}s` : `${m.name} ${m.reps}r`,
        )
        .join(" + ")}]`,
  );

describe("draft", () => {
  const start = fromDay(day);

  it("loads a day", () => {
    expect(lines(start)).toEqual(["3×[لانژ 12r]", "4×[پل باسن 45s]"]);
    expect(start[1]).toMatchObject({ key: "b2", rest_seconds: 60, notes: "کند" });
    expect(lines(run([], { type: "load", day }))).toEqual(lines(start));
  });

  it("adds a line copying the previous line's sets and amount", () => {
    const after = run(start, { type: "add", blockKey: "n1", movementKey: "nm1", movement: squat });
    expect(lines(after).at(-1)).toBe("4×[اسکوات 45s]");
    expect(after.at(-1)!.movements[0].load_kg).toBeNull();
  });

  it("adds a first line with the default prescription", () => {
    expect(
      lines(run([], { type: "add", blockKey: "n1", movementKey: "nm1", movement: squat })),
    ).toEqual(["3×[اسکوات 12r]"]);
  });

  it("sets a movement's weight in kg, and none at zero", () => {
    const kg = (value: number | null) =>
      run(start, { type: "kg", blockKey: "b1", movementKey: "bm1", kg: value })[0].movements[0]
        .load_kg;
    expect(kg(15)).toBe(15);
    expect(kg(17.5)).toBe(17.5);
    expect(kg(2.345)).toBe(2.35);
    expect(kg(900)).toBe(500);
    expect(kg(0)).toBeNull();
    expect(kg(0.001)).toBeNull();
    expect(kg(null)).toBeNull();
  });

  it("steps the weight by whole kilos, and below one is none", () => {
    expect(stepKg(null, 1)).toBe(1);
    expect(stepKg(17.5, 1)).toBe(18.5);
    expect(stepKg(1, -1)).toBeNull();
    expect(stepKg(null, -1)).toBeNull();
  });

  it("switches between reps and seconds", () => {
    const toSeconds = run(start, {
      type: "mode",
      blockKey: "b1",
      movementKey: "bm1",
      mode: "seconds",
    });
    expect(lines(toSeconds)[0]).toBe("3×[لانژ 30s]");
    const back = run(toSeconds, { type: "mode", blockKey: "b1", movementKey: "bm1", mode: "reps" });
    expect(lines(back)[0]).toBe("3×[لانژ 12r]");
    expect(run(start, { type: "mode", blockKey: "b1", movementKey: "bm1", mode: "reps" })).toEqual(
      start,
    );
  });

  it("steps reps by one and seconds by five, within bounds", () => {
    expect(
      lines(run(start, { type: "step", blockKey: "b1", movementKey: "bm1", delta: 1 }))[0],
    ).toBe("3×[لانژ 13r]");
    expect(
      lines(run(start, { type: "step", blockKey: "b2", movementKey: "bm2", delta: 1 }))[1],
    ).toBe("4×[پل باسن 50s]");
    const one = run(
      start,
      { type: "preset", blockKey: "b1", preset: { reps: 12, sets: 3 } },
      ...Array.from(
        { length: 20 },
        () => ({ type: "step", blockKey: "b1", movementKey: "bm1", delta: -1 }) as const,
      ),
    );
    expect(lines(one)[0]).toBe("3×[لانژ 1r]");
  });

  it("steps seconds by a minute past two minutes, and back by five below them", () => {
    const secondsAfter = (from: number, delta: 1 | -1) => {
      const seeded: Draft = start.map((b) =>
        b.key === "b2"
          ? { ...b, movements: b.movements.map((m) => ({ ...m, duration_seconds: from })) }
          : b,
      );
      const after = run(seeded, { type: "step", blockKey: "b2", movementKey: "bm2", delta });
      return after[1].movements[0].duration_seconds;
    };
    expect(secondsAfter(115, 1)).toBe(120);
    expect(secondsAfter(120, 1)).toBe(180);
    expect(secondsAfter(600, 1)).toBe(660);
    expect(secondsAfter(600, -1)).toBe(540);
    expect(secondsAfter(180, -1)).toBe(120);
    expect(secondsAfter(120, -1)).toBe(115);
    expect(secondsAfter(5, -1)).toBe(5);
    expect(secondsAfter(47, 1)).toBe(50);
    expect(secondsAfter(47, -1)).toBe(45);
    expect(secondsAfter(150, 1)).toBe(180);
    expect(secondsAfter(150, -1)).toBe(120);
    expect(secondsAfter(14400, 1)).toBe(14400);
  });

  it("offers the ten-minute single set, and recognizes it", () => {
    const preset = presets[5];
    const after = run(start, { type: "preset", blockKey: "b1", preset });
    expect(lines(after)[0]).toBe("1×[لانژ 600s]");
    expect(presetMatches(after[0], preset)).toBe(true);
    expect(presetMatches(after[0], presets[4])).toBe(false);
  });

  it("changes sets within 1-50", () => {
    expect(lines(run(start, { type: "sets", blockKey: "b1", delta: 1 }))[0]).toBe("4×[لانژ 12r]");
    const many = Array.from(
      { length: 5 },
      () => ({ type: "sets", blockKey: "b1", delta: -1 }) as const,
    );
    expect(run(start, ...many)[0].sets).toBe(1);
  });

  it("applies a preset to every movement of the line, and recognizes it", () => {
    const withSuperset = run(start, {
      type: "join",
      blockKey: "b1",
      movementKey: "nm1",
      movement: squat,
    });
    const preset = presets[4];
    const after = run(withSuperset, { type: "preset", blockKey: "b1", preset });
    expect(lines(after)[0]).toBe("3×[لانژ 30s + اسکوات 30s]");
    expect(presetMatches(after[0], preset)).toBe(true);
    expect(presetMatches(after[0], presets[0])).toBe(false);
  });

  it("makes a superset by joining a movement to a line, and breaks it apart again", () => {
    const joined = run(start, {
      type: "join",
      blockKey: "b1",
      movementKey: "nm1",
      movement: plank,
    });
    expect(lines(joined)).toEqual(["3×[لانژ 12r + پلانک 12r]", "4×[پل باسن 45s]"]);

    const broken = run(joined, { type: "split", blockKey: "b1", keys: ["n2"] });
    expect(lines(broken)).toEqual(["3×[لانژ 12r]", "3×[پلانک 12r]", "4×[پل باسن 45s]"]);
    expect(broken.map((b) => b.key)).toEqual(["b1", "n2", "b2"]);
  });

  it("removes one movement from a superset, and the line with its last movement", () => {
    const joined = run(start, {
      type: "join",
      blockKey: "b1",
      movementKey: "nm1",
      movement: plank,
    });
    expect(
      lines(run(joined, { type: "removeMovement", blockKey: "b1", movementKey: "bm1" }))[0],
    ).toBe("3×[پلانک 12r]");
    expect(
      lines(run(start, { type: "removeMovement", blockKey: "b1", movementKey: "bm1" })),
    ).toEqual(["4×[پل باسن 45s]"]);
  });

  it("moves a line past a superset whole", () => {
    const draft = run(
      start,
      { type: "join", blockKey: "b2", movementKey: "nm1", movement: plank },
      { type: "add", blockKey: "n3", movementKey: "nm3", movement: squat },
    );
    expect(lines(draft)).toEqual(["3×[لانژ 12r]", "4×[پل باسن 45s + پلانک 45s]", "4×[اسکوات 45s]"]);

    const down = run(draft, { type: "move", blockKey: "b1", delta: 1 });
    expect(lines(down)).toEqual(["4×[پل باسن 45s + پلانک 45s]", "3×[لانژ 12r]", "4×[اسکوات 45s]"]);
    const up = run(draft, { type: "move", blockKey: "n3", delta: -1 });
    expect(lines(up)).toEqual(["3×[لانژ 12r]", "4×[اسکوات 45s]", "4×[پل باسن 45s + پلانک 45s]"]);
    expect(run(draft, { type: "move", blockKey: "b1", delta: -1 })).toBe(draft);
  });

  it("deletes a line", () => {
    expect(lines(run(start, { type: "remove", blockKey: "b1" }))).toEqual(["4×[پل باسن 45s]"]);
  });

  it("turns into the request body, in order, without keys or names", () => {
    const draft = run(
      start,
      { type: "join", blockKey: "b1", movementKey: "nm1", movement: plank },
      { type: "move", blockKey: "b2", delta: -1 },
    );
    expect(toBody(draft)).toEqual([
      {
        sets: 4,
        rest_seconds: 60,
        notes: "کند",
        movements: [{ movement_id: bridge.id, reps: null, duration_seconds: 45, load_kg: 10 }],
      },
      {
        sets: 3,
        rest_seconds: null,
        notes: null,
        movements: [
          { movement_id: lunge.id, reps: 12, duration_seconds: null, load_kg: null },
          { movement_id: plank.id, reps: 12, duration_seconds: null, load_kg: null },
        ],
      },
    ]);
  });
});
