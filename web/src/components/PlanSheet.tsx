import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { fieldErrors } from "../api/client";
import { deletePlan, updatePlan } from "../api/plans";
import type { PlanSummary } from "../api/types";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import { ConfirmSheet } from "./ConfirmSheet";
import { JalaliDateField } from "./JalaliDateField";
import { TextField } from "./TextField";
import { msgTitleEmpty } from "../lib/formMessages";
import s from "./PlanSheet.module.css";

type PlanSheetProps = {
  open: boolean;
  onClose: () => void;
  plan: PlanSummary;
  onSaved: () => void;
};

// Behind a profile plan row's edit: the plan's title and start date.
export function PlanSheet({ open, onClose, plan, onSaved }: PlanSheetProps) {
  return (
    <BottomSheet open={open} onClose={onClose} title="برنامه">
      {/* Mounted only while open, so it starts from the saved values each time. */}
      {open && <PlanForm plan={plan} onSaved={onSaved} />}
    </BottomSheet>
  );
}

function PlanForm({ plan, onSaved }: Omit<PlanSheetProps, "open" | "onClose">) {
  const [title, setTitle] = useState(plan.title);
  const [startDate, setStartDate] = useState(plan.start_date);
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [titleError, setTitleError] = useState<string | null>(null);

  // PUT replaces every field, so the note the profile doesn't show is sent back unchanged.
  const save = useMutation({
    mutationFn: () =>
      updatePlan(plan.id, { title: title.trim(), start_date: startDate, note: plan.note }),
    onSuccess: onSaved,
  });
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
