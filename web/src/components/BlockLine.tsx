import type { ReactNode } from "react";
import { toPersianDigits } from "../lib/digits";
import { prescriptionText } from "../lib/prescription";
import s from "./BlockLine.module.css";

type LineMovement = {
  name: string;
  reps: number | null;
  duration_seconds: number | null;
  load: string | null;
};

type BlockLineProps = {
  index: number;
  block: { sets: number; movements: LineMovement[] };
};

export function DayHeading({ label, lines }: { label: string; lines: number }) {
  return (
    <div className={s.dh}>
      <b>روز {label}</b>
      <span>{toPersianDigits(lines)} حرکت</span>
    </div>
  );
}

// One numbered line: the movement, or a superset's movements one per row, and the prescription under it.
export function BlockLine({ index, block }: BlockLineProps) {
  const superset = block.movements.length > 1;
  return (
    <div className={s.mv}>
      <span className={s.num}>{toPersianDigits(index + 1)}</span>
      <BlockBody superset={superset}>
        {block.movements.map((m, j) => (
          <div key={j} className={s.nm}>
            {j > 0 && <i className={s.plus}>+</i>}
            {m.name}
            {j === 0 && superset && <span className={s.tag}>سوپرست</span>}
          </div>
        ))}
        <div className={s.rx}>{prescriptionText(block)}</div>
      </BlockBody>
    </div>
  );
}

function BlockBody({ superset, children }: { superset: boolean; children: ReactNode }) {
  return <div className={superset ? `${s.mb} ${s.ss}` : s.mb}>{children}</div>;
}
