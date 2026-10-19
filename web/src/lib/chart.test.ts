import { describe, expect, it } from "vitest";
import { chart, labelledIndexes, points } from "./chart";

const plotTop = chart.top;
const plotBottom = chart.height - chart.bottom;

describe("points", () => {
  it("puts a single check-in in the middle", () => {
    const [p] = points([70]);
    expect(p.x).toBe(chart.width / 2);
    expect(p.y).toBeCloseTo((plotTop + plotBottom) / 2);
  });

  it("draws equal values as a flat line through the middle", () => {
    const ps = points([70, 70, 70]);
    expect(new Set(ps.map((p) => p.y.toFixed(3))).size).toBe(1);
    expect(ps[0].y).toBeCloseTo((plotTop + plotBottom) / 2);
  });

  it("runs time right to left, with a falling line going down toward the left", () => {
    const ps = points([73.4, 72, 71.2, 70.4, 69.6]);
    expect(ps[0].x).toBe(chart.width - chart.padX);
    expect(ps.at(-1)!.x).toBe(chart.padX);
    for (let i = 1; i < ps.length; i++) {
      expect(ps[i].x).toBeLessThan(ps[i - 1].x);
      expect(ps[i].y).toBeGreaterThan(ps[i - 1].y);
    }
    expect(Math.min(...ps.map((p) => p.y))).toBeGreaterThan(plotTop);
    expect(Math.max(...ps.map((p) => p.y))).toBeLessThan(plotBottom);
  });

  it("returns nothing for no check-ins", () => {
    expect(points([])).toEqual([]);
  });
});

describe("labelledIndexes", () => {
  it("labels every point up to six", () => {
    expect(labelledIndexes(5)).toEqual([0, 1, 2, 3, 4]);
  });

  it("thins out longer histories and always labels the newest", () => {
    const picked = labelledIndexes(14);
    expect(picked.length).toBeLessThanOrEqual(6);
    expect(picked[0]).toBe(0);
    expect(picked.at(-1)).toBe(13);
  });
});
