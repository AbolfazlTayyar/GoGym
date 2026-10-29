import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router";
import { fieldErrors } from "../api/client";
import { queryKeys } from "../api/keys";
import {
  createMeasurement,
  deleteMeasurement,
  updateMeasurement,
  type MeasurementInput,
} from "../api/measurements";
import type { Measurement } from "../api/types";
import { parseNumber, toPersianDigits } from "../lib/digits";
import { focusNextOnEnter } from "../lib/focusNext";
import { msgNotANumber } from "../lib/formMessages";
import { todayISO } from "../lib/jalali";
import { formatValue, lastValue, metrics, type Metric } from "../lib/measurements";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import { ConfirmSheet } from "./ConfirmSheet";
import { TrashIcon } from "./icons";
import { JalaliDateField } from "./JalaliDateField";
import s from "./MeasurementSheet.module.css";
import { useSheet } from "./useSheet";

// The API's answer when no measurement is filled; it names every field, so it is shown once instead.
const msgNoneFilled = "دست‌کم یکی را پر کنید.";

type MeasurementSheetProps = {
  open: boolean;
  onClose: () => void;
  athleteId: string;
  // Oldest first, for each field's last value.
  measurements: Measurement[];
  // The check-in to edit; none for a new one.
  editing: Measurement | null;
};

export function MeasurementSheet(props: MeasurementSheetProps) {
  const { open, onClose, editing } = props;
  return (
    <BottomSheet open={open} onClose={onClose} title={editing ? "ویرایش اندازه" : "ثبت اندازه"}>
      {/* Mounted per opening, so a new check-in starts empty and an edit from its saved values. */}
      {open && <MeasurementForm key={editing?.id ?? "new"} {...props} />}
    </BottomSheet>
  );
}

function MeasurementForm({ onClose, athleteId, measurements, editing }: MeasurementSheetProps) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const confirm = useSheet("delete-measurement");
  const [date, setDate] = useState(editing?.date ?? todayISO());
  const [values, setValues] = useState<Record<Metric, string>>(
    () =>
      Object.fromEntries(
        metrics.map(({ key }) => [
          key,
          editing?.[key] == null ? "" : toPersianDigits(editing[key]),
        ]),
      ) as Record<Metric, string>,
  );
  const [typoFields, setTypoFields] = useState<Metric[]>([]);

  const done = async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.measurements() });
    onClose();
  };
  // Only the fields she filled are sent; the others go as null.
  const save = useMutation({
    mutationFn: (input: MeasurementInput) =>
      editing ? updateMeasurement(editing.id, input) : createMeasurement(athleteId, input),
    onSuccess: done,
  });

  const errors = fieldErrors(save.error);
  const noneFilled = metrics.every(({ key }) => errors[key] === msgNoneFilled);

  const submit = () => {
    const bad = metrics.filter(
      ({ key }) => values[key].trim() && parseNumber(values[key]) === null,
    );
    setTypoFields(bad.map((m) => m.key));
    if (bad.length) return;
    const input = { date } as MeasurementInput;
    for (const { key } of metrics) input[key] = parseNumber(values[key]);
    save.mutate(input);
  };

  return (
    <>
      <JalaliDateField label="تاریخ" value={date} onChange={setDate} error={errors.date} />
      <div className={s.mgrid} onKeyDown={focusNextOnEnter}>
        {metrics.map(({ key, label, unit }) => {
          const prev = lastValue(measurements, key, editing ?? undefined);
          const error = typoFields.includes(key)
            ? msgNotANumber
            : noneFilled
              ? undefined
              : errors[key];
          return (
            <label key={key} className={error ? `${s.mf} ${s.invalid}` : s.mf}>
              <span className={s.label}>
                <span>{label}</span>
                <span>{unit}</span>
              </span>
              <input
                inputMode="decimal"
                placeholder="—"
                value={values[key]}
                onChange={(e) => setValues((v) => ({ ...v, [key]: e.target.value }))}
                aria-label={label}
                aria-invalid={error ? true : undefined}
              />
              {error ? (
                <span className={s.error}>{error}</span>
              ) : (
                <span className={s.prev}>
                  {prev === null ? "قبلی —" : `قبلی ${formatValue(prev)}`}
                </span>
              )}
            </label>
          );
        })}
      </div>
      <p className={noneFilled ? `${s.hint} ${s.bad}` : s.hint}>
        فقط اندازه‌هایی را که گرفتید وارد کنید
      </p>
      <div className={s.acts}>
        {editing && (
          <button type="button" className={s.del} onClick={confirm.show}>
            <TrashIcon /> حذف
          </button>
        )}
        <BigButton onClick={submit} disabled={save.isPending}>
          ثبت
        </BigButton>
      </div>
      {save.isError && Object.keys(errors).length === 0 && (
        <p className={`${s.hint} ${s.bad}`} role="alert">
          {save.error.message}
        </p>
      )}
      {editing && (
        <DeleteMeasurementSheet
          open={confirm.open}
          onClose={confirm.close}
          measurement={editing}
          // The confirmation and the sheet under it: two history entries.
          onDeleted={() => navigate(-2)}
        />
      )}
    </>
  );
}

type DeleteMeasurementSheetProps = {
  open: boolean;
  onClose: () => void;
  measurement: Measurement;
  onDeleted: () => void;
};

// Asks before deleting a check-in, then deletes it and refreshes the history and the chart.
export function DeleteMeasurementSheet({
  open,
  onClose,
  measurement,
  onDeleted,
}: DeleteMeasurementSheetProps) {
  const queryClient = useQueryClient();
  const remove = useMutation({
    mutationFn: () => deleteMeasurement(measurement.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.measurements() });
      onDeleted();
    },
  });
  return (
    <ConfirmSheet
      open={open}
      onClose={onClose}
      title="حذف این اندازه‌گیری؟"
      text="این ثبت از تاریخچه و نمودار پاک می‌شود و برنمی‌گردد."
      confirmLabel="حذف"
      onConfirm={() => remove.mutate()}
      pending={remove.isPending}
      error={remove.isError ? remove.error.message : null}
    />
  );
}
