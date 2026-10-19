import type { ReactNode } from "react";
import s from "./PillTabs.module.css";

type PillTabsProps<T extends string> = {
  options: { value: T; label: ReactNode }[];
  value: T;
  onChange: (value: T) => void;
};

export function PillTabs<T extends string>({ options, value, onChange }: PillTabsProps<T>) {
  return (
    <div className={s.pills2} role="tablist">
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          role="tab"
          aria-selected={o.value === value}
          className={o.value === value ? s.on : undefined}
          onClick={() => onChange(o.value)}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
