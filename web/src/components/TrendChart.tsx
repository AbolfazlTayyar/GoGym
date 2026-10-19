import { chart, labelledIndexes, points } from "../lib/chart";
import { formatJalaliShort } from "../lib/jalali";
import { formatValue } from "../lib/measurements";
import s from "./TrendChart.module.css";

type TrendChartProps = {
  // Oldest first; only the check-ins that have this measurement.
  data: { date: string; value: number }[];
  unit: string;
};

// A hand-written SVG (docs/adr/0030): the line, its points, a soft fill under it and Jalali dates.
export function TrendChart({ data, unit }: TrendChartProps) {
  if (data.length === 0) {
    return (
      <div className={s.chart}>
        <p className={s.note}>این اندازه هنوز ثبت نشده است.</p>
      </div>
    );
  }

  const latest = data.at(-1)!;
  const change = latest.value - data[0].value;
  const ps = points(data.map((d) => d.value));
  const line = ps.map((p, i) => `${i ? "L" : "M"}${p.x.toFixed(1)} ${p.y.toFixed(1)}`).join(" ");
  const base = chart.height - chart.bottom;
  const labelled = new Set(labelledIndexes(data.length));
  const grid = [0, 1, 2].map((k) => chart.top + (k * (base - chart.top)) / 2);

  return (
    <div className={s.chart}>
      <div className={s.cv}>
        <b>{formatValue(latest.value)}</b>
        <span>{unit}</span>
        {data.length > 1 && (
          <span className={s.delta}>
            {change === 0
              ? `بدون تغییر از ${formatJalaliShort(data[0].date)}`
              : `${formatValue(Math.abs(change))} ${change < 0 ? "کمتر" : "بیشتر"} از ${formatJalaliShort(data[0].date)}`}
          </span>
        )}
      </div>
      <svg viewBox={`0 0 ${chart.width} ${chart.height}`} role="img" aria-label="نمودار تغییرات">
        <defs>
          <linearGradient id="trend-area" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stopColor="#FF0057" stopOpacity="0.22" />
            <stop offset="1" stopColor="#FF0057" stopOpacity="0" />
          </linearGradient>
        </defs>
        {grid.map((y) => (
          <line key={y} x1="0" x2={chart.width} y1={y} y2={y} stroke="#F7E8EC" />
        ))}
        {ps.length > 1 && (
          <>
            <path
              d={`${line} L${ps.at(-1)!.x} ${base} L${ps[0].x} ${base} Z`}
              fill="url(#trend-area)"
            />
            <path d={line} fill="none" stroke="#FF0057" strokeWidth="2.5" strokeLinejoin="round" />
          </>
        )}
        {ps.map((p, i) => {
          const newest = i === ps.length - 1;
          return (
            <circle
              key={i}
              cx={p.x}
              cy={p.y}
              r={newest ? 5.5 : 3.5}
              fill={newest ? "#FF0057" : "#fff"}
              stroke="#FF0057"
              strokeWidth="2"
            />
          );
        })}
        {ps.map(
          (p, i) =>
            labelled.has(i) && (
              <text key={i} x={p.x} y={chart.height - 8} textAnchor="middle">
                {formatJalaliShort(data[i].date)}
              </text>
            ),
        )}
      </svg>
    </div>
  );
}
