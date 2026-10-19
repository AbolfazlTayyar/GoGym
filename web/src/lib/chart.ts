// The measurement chart's geometry, from the board's SVG (docs/design/src/concept-8-2-athlete.html).

// Time runs right to left, like the text and the board. The design README asks to check this with the
// coach; flipping it is this one constant.
export const timeRunsRightToLeft = true;

export const chart = { width: 334, height: 118, padX: 22, top: 12, bottom: 26 };

// Space above and below the line, in the metric's own unit, so a flat line doesn't sit on the edge.
const valuePadding = 0.6;

export type Point = { x: number; y: number };

// points maps values, oldest first, onto the chart. A single check-in sits in the middle.
export function points(values: number[]): Point[] {
  const { width, height, padX, top, bottom } = chart;
  if (values.length === 0) return [];
  const lo = Math.min(...values) - valuePadding;
  const hi = Math.max(...values) + valuePadding;
  const plotHeight = height - top - bottom;
  const step = values.length > 1 ? (width - 2 * padX) / (values.length - 1) : 0;

  return values.map((v, i) => {
    const offset = values.length > 1 ? padX + i * step : width / 2;
    return {
      x: timeRunsRightToLeft ? width - offset : offset,
      y: top + ((hi - v) / (hi - lo)) * plotHeight,
    };
  });
}

// Which points get a date under them: all of them up to six, then every k-th and always the newest.
export function labelledIndexes(count: number, maxLabels = 6): number[] {
  if (count <= maxLabels) return Array.from({ length: count }, (_, i) => i);
  const every = Math.ceil(count / maxLabels);
  const picked = Array.from({ length: count }, (_, i) => i).filter((i) => i % every === 0);
  if (picked.at(-1) !== count - 1) {
    // The newest always shows; drop the one before it if the two would crowd.
    if (count - 1 - picked.at(-1)! < every) picked.pop();
    picked.push(count - 1);
  }
  return picked;
}
