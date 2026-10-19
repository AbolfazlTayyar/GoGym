import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { fieldErrors } from "../../api/client";
import { deletePlan, updatePlan } from "../../api/plans";
import type { PlanDetail } from "../../api/types";
import { BigButton } from "../../components/BigButton";
import { BottomSheet } from "../../components/BottomSheet";
import { ConfirmSheet } from "../../components/ConfirmSheet";
import { TrashIcon } from "../../components/icons";
import { JalaliDateField } from "../../components/JalaliDateField";
import { TextField } from "../../components/TextField";
import { useSheet } from "../../components/useSheet";
import s from "./Builder.module.css";

type PlanSheetProps = {
  open: boolean;
  onClose: () => void;
  plan: PlanDetail;
  onSaved: (plan: PlanDetail) => void;
  onDeleted: () => void;
};

// Behind the builder's title: the plan's title and start date, and deleting the plan.
export function PlanSheet({ open, onClose, plan, onSaved, onDeleted }: PlanSheetProps) {
  return (
    <BottomSheet open={open} onClose={onClose} title="برنامه">
      {/* Mounted only while open, so it starts from the saved values each time. */}
      {open && <PlanForm plan={plan} onSaved={onSaved} onDeleted={onDeleted} />}
    </BottomSheet>
  );
}

function PlanForm({ plan, onSaved, onDeleted }: Omit<PlanSheetProps, "open" | "onClose">) {
  const [title, setTitle] = useState(plan.title);
  const [startDate, setStartDate] = useState(plan.start_date);
  const confirm = useSheet("delete-plan");

  // PUT replaces every field, so the note the builder doesn't show is sent back unchanged.
  const save = useMutation({
    mutationFn: () =>
      updatePlan(plan.id, { title: title.trim(), start_date: startDate, note: plan.note }),
    onSuccess: onSaved,
  });
  const remove = useMutation({ mutationFn: () => deletePlan(plan.id), onSuccess: onDeleted });
  const errors = fieldErrors(save.error);

  return (
    <>
      <div className={s.sheetFields}>
        <TextField
          label="عنوان"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          error={errors.title}
        />
        <JalaliDateField
          label="تاریخ شروع"
          value={startDate}
          onChange={setStartDate}
          error={errors.start_date}
        />
      </div>
      <div className={s.sheetActs}>
        <BigButton onClick={() => save.mutate()} disabled={save.isPending}>
          ذخیره
        </BigButton>
        <BigButton outline onClick={confirm.show}>
          <TrashIcon /> حذف برنامه
        </BigButton>
      </div>
      {save.isError && Object.keys(errors).length === 0 && (
        <p className={s.lineError} role="alert">
          {save.error.message}
        </p>
      )}
      <ConfirmSheet
        open={confirm.open}
        onClose={confirm.close}
        title={`حذف «${plan.title}»؟`}
        text="همه روزها و حرکت‌های این برنامه پاک می‌شوند و برنمی‌گردند."
        confirmLabel="حذف برنامه"
        onConfirm={() => remove.mutate()}
        pending={remove.isPending}
        error={remove.isError ? remove.error.message : null}
      />
    </>
  );
}
