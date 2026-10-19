import s from "./Segmented.module.css";

type SegmentedProps<T extends string> = {
  options: { value: T; label: string }[];
  value: T | null;
  onChange: (value: T | null) => void;
  // A form field, under a label; otherwise the builder's smaller inline one.
  field?: boolean;
  // Tapping the chosen option again clears it, for an optional field.
  clearable?: boolean;
  label?: string;
};

export function Segmented<T extends string>({
  options,
  value,
  onChange,
  field,
  clearable,
  label,
}: SegmentedProps<T>) {
  return (
    <div className={field ? `${s.seg} ${s.field}` : s.seg} role="radiogroup" aria-label={label}>
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          role="radio"
          aria-checked={o.value === value}
          className={o.value === value ? s.on : undefined}
          onClick={() => onChange(clearable && o.value === value ? null : o.value)}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
