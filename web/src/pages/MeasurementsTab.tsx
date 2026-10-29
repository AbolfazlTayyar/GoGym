import { useState } from "react";
import { useSearchParams } from "react-router";
import type { Measurement } from "../api/types";
import { ChipRow } from "../components/ChipRow";
import { EmptyState } from "../components/EmptyState";
import { ChevLIcon } from "../components/icons";
import { SwipeRow } from "../components/SwipeRow";
import { TrendChart } from "../components/TrendChart";
import { formatJalali } from "../lib/jalali";
import { formatValue, metrics, series, type Metric } from "../lib/measurements";
import s from "./ProfilePage.module.css";

const metricParam = "metric";

type MeasurementsTabProps = {
  // Oldest first, as the API returns them.
  measurements: Measurement[];
  onOpen: (m: Measurement) => void;
  onDelete: (m: Measurement) => void;
};

// One measurement at a time as a trend, then every check-in newest first.
export function MeasurementsTab({ measurements, onOpen, onDelete }: MeasurementsTabProps) {
  const [params, setParams] = useSearchParams();
  const [swipedId, setSwipedId] = useState<string | null>(null);
  const metric = metrics.find((m) => m.key === params.get(metricParam)) ?? metrics[0];

  if (measurements.length === 0) {
    return <EmptyState title="هنوز اندازه‌ای ثبت نشده" text="با + ثبت اندازه اولین را ثبت کنید." />;
  }

  return (
    <>
      <ChipRow
        options={metrics.map((m) => ({ value: m.key, label: m.label }))}
        value={metric.key}
        onChange={(key: Metric) =>
          setParams(
            (prev) => {
              const next = new URLSearchParams(prev);
              next.set(metricParam, key);
              return next;
            },
            { replace: true },
          )
        }
      />
      <TrendChart data={series(measurements, metric.key)} unit={metric.unit} />
      <div className={s.lbl}>تاریخچه</div>
      <div className={s.hist}>
        {[...measurements].reverse().map((m) => (
          <SwipeRow
            key={m.id}
            className={s.hrow}
            open={swipedId === m.id}
            onOpenChange={(o) => setSwipedId(o ? m.id : null)}
            onEdit={() => {
              setSwipedId(null);
              onOpen(m);
            }}
            onDelete={() => {
              setSwipedId(null);
              onDelete(m);
            }}
          >
            <button type="button" className={s.hr} onClick={() => onOpen(m)}>
              <div className={s.nm}>
                <b>{formatJalali(m.date)}</b>
                <span>
                  {metrics
                    .filter(({ key }) => m[key] !== null)
                    .map(({ key, label }) => `${label} ${formatValue(m[key]!)}`)
                    .join(" · ")}
                </span>
              </div>
              <span className={s.go}>
                <ChevLIcon />
              </span>
            </button>
          </SwipeRow>
        ))}
      </div>
    </>
  );
}
