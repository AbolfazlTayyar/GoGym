import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { getAttendance, trackPlan, untrackPlan } from "../api/attendance";
import { fieldErrors } from "../api/client";
import { queryKeys } from "../api/keys";
import { deletePlan, updatePlan } from "../api/plans";
import type { Attendance, PlanSummary } from "../api/types";
import { toPersianDigits } from "../lib/digits";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import { ConfirmSheet } from "./ConfirmSheet";
import { JalaliDateField } from "./JalaliDateField";
import { SessionToggle } from "./SessionToggle";
import { TextAreaField, TextField } from "./TextField";
import { msgTitleEmpty } from "../lib/formMessages";
import s from "./PlanSheet.module.css";

type PlanSheetProps = {
  open: boolean;
  onClose: () => void;
  plan: PlanSummary;
  onSaved: () => void;
};

// Behind a profile plan row's edit: the plan's title, start date, note and session tracking.
export function PlanSheet({ open, onClose, plan, onSaved }: PlanSheetProps) {
  const attendance = useQuery({
    queryKey: queryKeys.attendance("plan", plan.id),
    queryFn: () => getAttendance(plan.id),
    enabled: open,
  });
  return (
    <BottomSheet open={open} onClose={onClose} title="برنامه">
      {/* Mounted only while open and loaded, so it starts from the saved values each time. */}
      {open && attendance.data && (
        <PlanForm plan={plan} attendance={attendance.data} onSaved={onSaved} />
      )}
      {open && attendance.isError && (
        <p className={s.error} role="alert">
          {attendance.error.message}
        </p>
      )}
    </BottomSheet>
  );
}

type PlanFormProps = Omit<PlanSheetProps, "open" | "onClose"> & { attendance: Attendance };

function PlanForm({ plan, attendance, onSaved }: PlanFormProps) {
  const [tracked, setTracked] = useState(attendance.tracked);
  const [title, setTitle] = useState(plan.title);
  const [startDate, setStartDate] = useState(plan.start_date);
  const [note, setNote] = useState(plan.note ?? "");
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [titleError, setTitleError] = useState<string | null>(null);

  const save = useMutation({
    mutationFn: async () => {
      await updatePlan(plan.id, {
        title: title.trim(),
        start_date: startDate,
        note: note.trim() || null,
      });
      if (tracked !== attendance.tracked) {
        await (tracked ? trackPlan(plan.id) : untrackPlan(plan.id));
      }
    },
    onSuccess: onSaved,
  });
  // The warning stands in for a confirm: saving with it showing deletes the visits.
  const visits = attendance.dates.length;
  const untrackWarning =
    attendance.tracked && !tracked && visits > 0
      ? `با ذخیره، ${toPersianDigits(visits)} حضور ثبت‌شده پاک می‌شود.`
      : null;
  const errors = fieldErrors(save.error);

  const submit = () => {
    if (!title.trim()) {
      setTitleError(msgTitleEmpty);
      save.reset();
      return;
    }
    setTitleError(null);
    save.mutate();
  };

  return (
    <>
      <div className={s.fields}>
        <TextField
          label="عنوان"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          error={titleError ?? errors.title}
        />
        <JalaliDateField
          label="تاریخ شروع"
          value={startDate}
          onChange={setStartDate}
          error={errors.start_date}
        />
        <SessionToggle on={tracked} onChange={setTracked} warning={untrackWarning} />
        <TextAreaField
          label="یادداشت"
          optional
          value={note}
          onChange={(e) => setNote(e.target.value)}
          error={errors.note}
        />
      </div>
      <BigButton onClick={submit} disabled={save.isPending}>
        ذخیره
      </BigButton>
      {save.isError && Object.keys(errors).length === 0 && (
        <p className={s.error} role="alert">
          {save.error.message}
        </p>
      )}
    </>
  );
}

type DeletePlanSheetProps = {
  open: boolean;
  onClose: () => void;
  plan: PlanSummary;
  onDeleted: () => void;
};

// Asks before deleting the plan, then deletes it.
export function DeletePlanSheet({ open, onClose, plan, onDeleted }: DeletePlanSheetProps) {
  const remove = useMutation({ mutationFn: () => deletePlan(plan.id), onSuccess: onDeleted });
  return (
    <ConfirmSheet
      open={open}
      onClose={onClose}
      title={`حذف «${plan.title}»؟`}
      text="همه روزها و حرکت‌های این برنامه پاک می‌شوند و برنمی‌گردند."
      confirmLabel="حذف برنامه"
      onConfirm={() => remove.mutate()}
      pending={remove.isPending}
      error={remove.isError ? remove.error.message : null}
    />
  );
}
