import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { addVisit, deleteVisit } from "../api/attendance";
import { fieldErrors } from "../api/client";
import { queryKeys } from "../api/keys";
import type { Attendance } from "../api/types";
import { msgSessionsDone, sessionsUsed, visitCountText } from "../lib/attendance";
import { toPersianDigits } from "../lib/digits";
import { formatJalali, todayISO } from "../lib/jalali";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import { ConfirmSheet } from "./ConfirmSheet";
import { JalaliDateField } from "./JalaliDateField";
import { SwipeRow } from "./SwipeRow";
import { useSheet } from "./useSheet";
import s from "./AttendanceSheet.module.css";

type AttendanceSheetProps = {
  open: boolean;
  onClose: () => void;
  planId: string;
  attendance: Attendance;
};

// The plan's visits: delete one recorded by mistake, or add a day the coach forgot to tap.
export function AttendanceSheet({ open, onClose, planId, attendance }: AttendanceSheetProps) {
  const queryClient = useQueryClient();
  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.attendance() });
  const confirm = useSheet("delete-visit");
  const [deleting, setDeleting] = useState<string | null>(null);
  const [swiped, setSwiped] = useState<string | null>(null);
  const [date, setDate] = useState(() => todayISO());

  const add = useMutation({ mutationFn: () => addVisit(planId, date), onSuccess: refresh });
  const remove = useMutation({
    mutationFn: (d: string) => deleteVisit(planId, d),
    onSuccess: async () => {
      await refresh();
      confirm.close();
    },
  });
  const addError = fieldErrors(add.error).date ?? (add.isError ? add.error.message : undefined);
  const done = sessionsUsed(attendance.dates.length, attendance.sessions);

  return (
    <BottomSheet
      open={open}
      onClose={onClose}
      title={`حضور · ${visitCountText(attendance.dates.length, attendance.sessions)} جلسه`}
    >
      {attendance.dates.length === 0 ? (
        <p className={s.note}>هنوز حضوری ثبت نشده است.</p>
      ) : (
        <div className={s.list}>
          {attendance.dates.map((d, i) => (
            <SwipeRow
              key={d}
              className={s.vrow}
              open={swiped === d}
              onOpenChange={(o) => setSwiped(o ? d : null)}
              onDelete={() => {
                setSwiped(null);
                setDeleting(d);
                remove.reset();
                confirm.show();
              }}
            >
              <div className={s.visit}>
                <span className={s.n}>{toPersianDigits(i + 1)}</span>
                <span className={s.date}>{formatJalali(d)}</span>
              </div>
            </SwipeRow>
          ))}
        </div>
      )}
      {done ? (
        <p className={s.done}>{msgSessionsDone}؛ با برنامه بعدی دوباره شمرده می‌شود.</p>
      ) : (
        <div className={s.add}>
          <JalaliDateField
            label="تاریخ حضور"
            value={date}
            onChange={(d) => {
              setDate(d);
              add.reset();
            }}
            max={todayISO()}
            error={addError}
          />
          <BigButton onClick={() => add.mutate()} disabled={add.isPending}>
            ثبت حضور
          </BigButton>
        </div>
      )}
      {deleting && (
        <ConfirmSheet
          open={confirm.open}
          onClose={confirm.close}
          title={`حذف حضور ${formatJalali(deleting)}؟`}
          confirmLabel="حذف"
          onConfirm={() => remove.mutate(deleting)}
          pending={remove.isPending}
          error={remove.isError ? remove.error.message : null}
        />
      )}
    </BottomSheet>
  );
}
