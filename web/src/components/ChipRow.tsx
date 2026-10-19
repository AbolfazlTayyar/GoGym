import type { ReactNode } from "react";
import s from "./ChipRow.module.css";

type ChipRowProps<T extends string> = {
  options: { value: T; label: ReactNode }[];
  value: T | null;
  onChange: (value: T) => void;
  // Inside a sheet or a field rather than under a header.
  inline?: boolean;
  // A last chip for adding a choice, such as "+ دسته جدید".
  trailing?: { label: string; onClick: () => void };
};

export function ChipRow<T extends string>({
  options,
  value,
  onChange,
  inline,
  trailing,
}: ChipRowProps<T>) {
  return (
    <div className={inline ? `${s.chiprow} ${s.inline}` : s.chiprow}>
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          aria-pressed={o.value === value}
          className={o.value === value ? s.on : undefined}
          onClick={() => onChange(o.value)}
        >
          {o.label}
        </button>
      ))}
      {trailing && (
        <button type="button" className={s.new} onClick={trailing.onClick}>
          {trailing.label}
        </button>
      )}
    </div>
  );
}
