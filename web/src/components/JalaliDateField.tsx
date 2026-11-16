import { useState } from "react";
import { toPersianDigits } from "../lib/digits";
import {
  formatJalali,
  isoToJalali,
  jalaliMonthLength,
  jalaliMonthNames,
  jalaliToISO,
  jalaliWeekdayIndex,
  jalaliWeekdayNames,
  type JalaliDate,
} from "../lib/jalali";
import { BottomSheet } from "./BottomSheet";
import f from "./Field.module.css";
import { CalendarIcon } from "./icons";
import s from "./JalaliDateField.module.css";
import { useSheet } from "./useSheet";

type JalaliDateFieldProps = {
  label: string;
  // ISO date (YYYY-MM-DD); both forms that use this field pre-fill it.
  value: string;
  onChange: (iso: string) => void;
  error?: string;
  // The last ISO date that can be picked; later days are disabled.
  max?: string;
};

export function JalaliDateField({ label, value, onChange, error, max }: JalaliDateFieldProps) {
  const sheet = useSheet(`date:${label}`);

  return (
    <div className={f.fl}>
      <span className={f.label}>{label}</span>
      <button type="button" className={error ? `${f.in} ${f.invalid}` : f.in} onClick={sheet.show}>
        <span>{formatJalali(value)}</span>
        <CalendarIcon />
      </button>
      {error && <span className={f.error}>{error}</span>}
      <BottomSheet open={sheet.open} onClose={sheet.close} title={label}>
        <Picker
          initial={isoToJalali(value)}
          max={max}
          onPick={(iso) => {
            onChange(iso);
            sheet.close();
          }}
        />
      </BottomSheet>
    </div>
  );
}

// Year and month change in place; tapping a day picks the date and closes the sheet.
function Picker({
  initial,
  max,
  onPick,
}: {
  initial: JalaliDate;
  max?: string;
  onPick: (iso: string) => void;
}) {
  const [jy, setYear] = useState(initial.jy);
  const [jm, setMonth] = useState(initial.jm);
  const length = jalaliMonthLength(jy, jm);
  const isPicked = (jd: number) => jy === initial.jy && jm === initial.jm && jd === initial.jd;
  const firstWeekday = jalaliWeekdayIndex(jalaliToISO({ jy, jm, jd: 1 }));

  return (
    <>
      <div className={s.year}>
        <button
          type="button"
          className={s.step}
          aria-label="سال بعد"
          onClick={() => setYear(jy + 1)}
        >
          +
        </button>
        <b>{toPersianDigits(jy)}</b>
        <button
          type="button"
          className={s.step}
          aria-label="سال قبل"
          onClick={() => setYear(jy - 1)}
        >
          −
        </button>
      </div>
      <div className={s.months}>
        {jalaliMonthNames.map((name, i) => (
          <button
            key={name}
            type="button"
            className={jm === i + 1 ? s.on : undefined}
            onClick={() => setMonth(i + 1)}
          >
            {name}
          </button>
        ))}
      </div>
      <div className={s.weekdays}>
        {jalaliWeekdayNames.map((name) => (
          <span key={name}>{name}</span>
        ))}
      </div>
      <div className={s.days}>
        {Array.from({ length }, (_, i) => i + 1).map((jd) => {
          const iso = jalaliToISO({ jy, jm, jd });
          return (
            <button
              key={jd}
              type="button"
              className={isPicked(jd) ? s.on : undefined}
              style={jd === 1 ? { gridColumnStart: firstWeekday + 1 } : undefined}
              // ISO dates compare as strings.
              disabled={max !== undefined && iso > max}
              onClick={() => onPick(iso)}
            >
              {toPersianDigits(jd)}
            </button>
          );
        })}
      </div>
    </>
  );
}
