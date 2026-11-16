import { toPersianDigits } from "../lib/digits";
import s from "./SessionToggle.module.css";

// What every gym sells to an athlete who comes every other day; the API returns it as `sessions` too.
export const sessionsPerPlan = 12;

type SessionToggleProps = {
  on: boolean;
  onChange: (on: boolean) => void;
  // Shown under the switch, e.g. what turning it off deletes.
  warning?: string | null;
};

// Turns on visit counting for a plan whose athlete pays per 12 sessions.
export function SessionToggle({ on, onChange, warning }: SessionToggleProps) {
  return (
    <div>
      <button
        type="button"
        role="switch"
        aria-checked={on}
        className={s.row}
        onClick={() => onChange(!on)}
      >
        <span className={s.text}>
          <b>یک روز در میان</b>
          <span>شمارش {toPersianDigits(sessionsPerPlan)} جلسه، با ثبت هر حضور</span>
        </span>
        <span className={on ? `${s.track} ${s.on}` : s.track} />
      </button>
      {warning && (
        <p className={s.warning} role="alert">
          {warning}
        </p>
      )}
    </div>
  );
}
