import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router";
import { ApiError, fieldErrors } from "../api/client";
import { ErrorCode } from "../api/codes";
import { queryKeys } from "../api/keys";
import { createMovement, deleteMovement, updateMovement } from "../api/movements";
import type { Movement } from "../api/types";
import { focusNextOnEnter } from "../lib/focusNext";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import { ChipRow } from "./ChipRow";
import { ConfirmSheet } from "./ConfirmSheet";
import f from "./Field.module.css";
import { LockIcon, TrashIcon } from "./icons";
import s from "./MovementSheet.module.css";
import { TextAreaField, TextField } from "./TextField";
import { useSheet } from "./useSheet";

type MovementSheetProps = {
  open: boolean;
  onClose: () => void;
  // The movement to show or edit; none for a new one.
  movement: Movement | null;
  // The categories in use, offered as chips.
  categories: string[];
};

export function MovementSheet(props: MovementSheetProps) {
  const { open, onClose, movement } = props;
  const universal = movement !== null && movement.coach_id === null;
  const title = movement === null ? "حرکت جدید" : universal ? "حرکت پیش‌فرض" : "ویرایش حرکت";
  return (
    <BottomSheet open={open} onClose={onClose} title={title}>
      {open && <MovementForm key={movement?.id ?? "new"} {...props} universal={universal} />}
    </BottomSheet>
  );
}

function MovementForm({
  onClose,
  movement,
  categories,
  universal,
}: MovementSheetProps & { universal: boolean }) {
  const navigate = useNavigate();
  const confirm = useSheet("delete-movement");
  const [name, setName] = useState(movement?.name ?? "");
  const [category, setCategory] = useState<string | null>(movement?.category ?? null);
  const [newCategory, setNewCategory] = useState<string | null>(null);
  const [description, setDescription] = useState(movement?.description ?? "");
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [nameError, setNameError] = useState<string | null>(null);

  const { refresh, onError } = useRefreshMovements();

  const save = useMutation({
    mutationFn: () => {
      const input = {
        name: name.trim(),
        category: (newCategory ?? category)?.trim() || null,
        description: description.trim() || null,
      };
      return movement ? updateMovement(movement.id, input) : createMovement(input);
    },
    onSuccess: async () => {
      await refresh();
      onClose();
    },
    onError,
  });
  const errors = fieldErrors(save.error);
  const submit = () => {
    if (!name.trim()) {
      setNameError("نام حرکت را وارد کنید.");
      save.reset();
      return;
    }
    setNameError(null);
    save.mutate();
  };
  const options = [...new Set([...categories, ...(category ? [category] : [])])];

  if (universal && movement) {
    return (
      <div className={s.fields}>
        <TextField label="نام" value={movement.name} readOnly />
        {movement.category && (
          <div className={f.fl}>
            <span className={f.label}>دسته‌بندی</span>
            <ChipRow
              inline
              options={[{ value: movement.category, label: movement.category }]}
              value={movement.category}
              onChange={() => {}}
            />
          </div>
        )}
        {movement.description && (
          <TextAreaField label="توضیح" value={movement.description} readOnly />
        )}
        <p className={s.locked}>
          <LockIcon /> حرکت پیش‌فرض است: در هر برنامه‌ای قابل استفاده است، ولی ویرایش یا حذف
          نمی‌شود.
        </p>
      </div>
    );
  }

  return (
    <>
      <div className={s.fields} onKeyDown={focusNextOnEnter}>
        <TextField
          label="نام"
          value={name}
          onChange={(e) => setName(e.target.value)}
          error={nameError ?? errors.name}
        />
        <div className={f.fl}>
          <span className={f.label}>
            دسته‌بندی <i className={f.optional}>· اختیاری</i>
          </span>
          <ChipRow
            inline
            options={options.map((c) => ({ value: c, label: c }))}
            value={newCategory === null ? category : null}
            onChange={(c) => {
              setNewCategory(null);
              // Tapping the chosen category again leaves the movement without one.
              setCategory(c === category ? null : c);
            }}
            trailing={{ label: "+ دسته جدید", onClick: () => setNewCategory(newCategory ?? "") }}
          />
          {errors.category && <span className={f.error}>{errors.category}</span>}
        </div>
        {newCategory !== null && (
          <TextField
            label="دسته جدید"
            value={newCategory}
            onChange={(e) => setNewCategory(e.target.value)}
            autoFocus
          />
        )}
        <TextAreaField
          label="توضیح"
          optional
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          error={errors.description}
        />
      </div>
      <div className={s.acts}>
        {movement && (
          <button type="button" className={s.del} onClick={confirm.show}>
            <TrashIcon /> حذف
          </button>
        )}
        <BigButton onClick={submit} disabled={save.isPending}>
          ذخیره
        </BigButton>
      </div>
      {save.isError && Object.keys(errors).length === 0 && (
        <p className={s.error} role="alert">
          {save.error.message}
        </p>
      )}
      {movement && (
        <DeleteMovementSheet
          open={confirm.open}
          onClose={confirm.close}
          movement={movement}
          // The confirmation and the sheet under it: two history entries.
          onDeleted={() => navigate(-2)}
        />
      )}
    </>
  );
}

type DeleteMovementSheetProps = {
  open: boolean;
  onClose: () => void;
  movement: Movement;
  onDeleted: () => void;
};

// Asks before deleting one of the coach's movements, then deletes it and refreshes the lists.
export function DeleteMovementSheet({
  open,
  onClose,
  movement,
  onDeleted,
}: DeleteMovementSheetProps) {
  const { refresh, onError } = useRefreshMovements();
  const remove = useMutation({
    mutationFn: () => deleteMovement(movement.id),
    onSuccess: async () => {
      await refresh();
      onDeleted();
    },
    onError,
  });
  return (
    <ConfirmSheet
      open={open}
      onClose={onClose}
      title={`حذف «${movement.name}»؟`}
      text="از فهرست حرکات شما حذف می‌شود، ولی برنامه‌هایی که از آن استفاده کرده‌اند همچنان آن را نشان می‌دهند."
      confirmLabel="حذف"
      onConfirm={() => remove.mutate()}
      pending={remove.isPending}
      error={remove.isError ? remove.error.message : null}
    />
  );
}

function useRefreshMovements() {
  const queryClient = useQueryClient();
  // The builder's search reads the same query, so it sees the change too.
  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.movements() });
  // A 403 means the list showed a built-in movement as the coach's own; refreshing puts it right.
  const onError = (err: Error) => {
    if (err instanceof ApiError && err.code === ErrorCode.forbidden) refresh();
  };
  return { refresh, onError };
}
