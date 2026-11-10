import { useState, type Dispatch, type RefObject } from "react";
import { TrashIcon, XIcon } from "../../components/icons";
import { Segmented } from "../../components/Segmented";
import { parseNumber, toPersianDigits } from "../../lib/digits";
import { amountText } from "../../lib/prescription";
import s from "./Builder.module.css";
import {
  modeOf,
  newKey,
  presetMatches,
  presets,
  stepKg,
  type Action,
  type DraftBlock,
  type DraftMovement,
  type Mode,
} from "./draft";

type OpenLineProps = {
  ref: RefObject<HTMLDivElement | null>;
  index: number;
  block: DraftBlock;
  isFirst: boolean;
  isLast: boolean;
  joining: boolean;
  dispatch: Dispatch<Action>;
  onClose: () => void;
  onJoin: () => void;
  error?: string;
};

const modeLabels: { value: Mode; label: string }[] = [
  { value: "reps", label: "تکرار" },
  { value: "seconds", label: "ثانیه" },
];

const presetLabel = (p: (typeof presets)[number]) =>
  "seconds" in p
    ? `${amountText({ reps: null, duration_seconds: p.seconds, load_kg: null })} × ${toPersianDigits(p.sets)}`
    : `${toPersianDigits(p.reps)} × ${toPersianDigits(p.sets)}`;

// A line opened in place: everything is a tap (type, +/-, presets, superset, move, delete), and a weight can
// also be typed, since 40 kg is too many taps.
// Rest is hidden: her plans never prescribe it (docs/design/README.md).
export function OpenLine(props: OpenLineProps) {
  const { ref, index, block, isFirst, isLast, joining, dispatch, onClose, onJoin, error } = props;
  const superset = block.movements.length > 1;
  const blockKey = block.key;

  return (
    <div ref={ref} className={s.open}>
      <div className={s.oh}>
        <span className={s.num}>{toPersianDigits(index + 1)}</span>
        <button type="button" className={s.ohTitle} onClick={onClose}>
          {superset ? "سوپرست" : block.movements[0].name}
        </button>
        <button
          type="button"
          className={s.del}
          aria-label="حذف خط"
          onClick={() => dispatch({ type: "remove", blockKey })}
        >
          <TrashIcon />
        </button>
      </div>
      {block.movements.map((m) => (
        <div key={m.key} className={superset ? s.part : undefined}>
          {superset && (
            <div className={s.partHead}>
              <span>{m.name}</span>
              <button
                type="button"
                className={s.del}
                aria-label={`برداشتن ${m.name} از سوپرست`}
                onClick={() => dispatch({ type: "removeMovement", blockKey, movementKey: m.key })}
              >
                <XIcon width={16} height={16} />
              </button>
            </div>
          )}
          <Segmented
            label="نوع"
            options={modeLabels}
            value={modeOf(m)}
            onChange={(mode) =>
              mode && dispatch({ type: "mode", blockKey, movementKey: m.key, mode })
            }
          />
          <div className={s.nums}>
            <Amount movement={m} blockKey={blockKey} dispatch={dispatch} />
            <Weight movement={m} blockKey={blockKey} dispatch={dispatch} />
          </div>
        </div>
      ))}
      <div className={s.nums}>
        <div>
          <label>ست</label>
          <Stepper
            value={block.sets}
            label="ست"
            onStep={(delta) => dispatch({ type: "sets", blockKey, delta })}
          />
        </div>
      </div>
      <div className={s.presets}>
        {presets.map((p, i) => (
          <button
            key={i}
            type="button"
            className={presetMatches(block, p) ? s.on : undefined}
            onClick={() => dispatch({ type: "preset", blockKey, preset: p })}
          >
            {presetLabel(p)}
          </button>
        ))}
      </div>
      <div className={s.links}>
        <button type="button" className={joining ? s.joining : undefined} onClick={onJoin}>
          + سوپرست
        </button>
        {superset && (
          <button
            type="button"
            onClick={() =>
              dispatch({
                type: "split",
                blockKey,
                keys: block.movements.slice(1).map(() => newKey()),
              })
            }
          >
            جدا کردن
          </button>
        )}
        <button
          type="button"
          disabled={isFirst}
          onClick={() => dispatch({ type: "move", blockKey, delta: -1 })}
        >
          ↑ بالا
        </button>
        <button
          type="button"
          disabled={isLast}
          onClick={() => dispatch({ type: "move", blockKey, delta: 1 })}
        >
          ↓ پایین
        </button>
      </div>
      {error && (
        <p className={s.lineError} role="alert">
          {error}
        </p>
      )}
    </div>
  );
}

function Amount({
  movement,
  blockKey,
  dispatch,
}: {
  movement: DraftMovement;
  blockKey: string;
  dispatch: Dispatch<Action>;
}) {
  const seconds = modeOf(movement) === "seconds";
  const label = seconds ? "ثانیه" : "تکرار";
  return (
    <div>
      <label>{label}</label>
      <Stepper
        value={(seconds ? movement.duration_seconds : movement.reps) ?? 0}
        label={label}
        onStep={(delta) => dispatch({ type: "step", blockKey, movementKey: movement.key, delta })}
      />
    </div>
  );
}

// The weight in kg; empty is no weight. The number is an input: the steppers' value until she types in it.
function Weight({
  movement,
  blockKey,
  dispatch,
}: {
  movement: DraftMovement;
  blockKey: string;
  dispatch: Dispatch<Action>;
}) {
  const kg = movement.load_kg;
  // What she is typing, held while the field has focus so "17." isn't redrawn as "17".
  const [typing, setTyping] = useState<string | null>(null);
  const set = (value: number | null) =>
    dispatch({ type: "kg", blockKey, movementKey: movement.key, kg: value });
  const label = "وزنه (کیلو)";
  return (
    <div>
      <label htmlFor={`kg-${movement.key}`}>{label}</label>
      <div className={s.ctl}>
        <button
          type="button"
          className={s.p}
          aria-label={`${label} بیشتر`}
          onClick={() => set(stepKg(kg, 1))}
        >
          +
        </button>
        <input
          id={`kg-${movement.key}`}
          inputMode="decimal"
          enterKeyHint="done"
          autoComplete="off"
          placeholder="—"
          value={typing ?? (kg === null ? "" : toPersianDigits(kg))}
          onFocus={(e) => setTyping(e.currentTarget.value)}
          onBlur={() => setTyping(null)}
          onChange={(e) => {
            const text = e.currentTarget.value;
            setTyping(text);
            if (text.trim() === "") set(null);
            else {
              const value = parseNumber(text);
              if (value !== null) set(value);
            }
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") e.currentTarget.blur();
          }}
        />
        <button type="button" aria-label={`${label} کمتر`} onClick={() => set(stepKg(kg, -1))}>
          −
        </button>
      </div>
    </div>
  );
}

// blush.css .ctl: + on the right (pink), the value, − on the left, as the board draws it in RTL.
function Stepper({
  value,
  label,
  onStep,
}: {
  value: number;
  label: string;
  onStep: (delta: 1 | -1) => void;
}) {
  return (
    <div className={s.ctl}>
      <button type="button" className={s.p} aria-label={`${label} بیشتر`} onClick={() => onStep(1)}>
        +
      </button>
      <b aria-live="polite">{toPersianDigits(value)}</b>
      <button type="button" aria-label={`${label} کمتر`} onClick={() => onStep(-1)}>
        −
      </button>
    </div>
  );
}
