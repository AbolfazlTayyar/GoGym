import type { ReactNode } from "react";
import s from "./DayPills.module.css";

type DayPillsProps = {
  labels: string[];
  current: string | null;
  onSelect: (label: string) => void;
  // The builder's extra pill after the days, such as "+".
  trailing?: ReactNode;
};

export function DayPills({ labels, current, onSelect, trailing }: DayPillsProps) {
  return (
    <div className={s.days} role="tablist">
      {labels.map((label) => (
        <button
          key={label}
          type="button"
          role="tab"
          aria-selected={label === current}
          className={label === current ? s.on : undefined}
          onClick={() => onSelect(label)}
        >
          {label}
        </button>
      ))}
      {trailing}
    </div>
  );
}
