import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { ApiError, fieldErrors } from "../api/client";
import { ErrorCode } from "../api/codes";
import { queryKeys } from "../api/keys";
import { deleteMovementCategory, renameMovementCategory } from "../api/movementCategories";
import type { Movement } from "../api/types";
import { normalizeCategory } from "../lib/category";
import { toPersianDigits } from "../lib/digits";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import s from "./CategoriesSheet.module.css";
import { ConfirmSheet } from "./ConfirmSheet";
import { ChevLIcon, LockIcon, TrashIcon } from "./icons";
import p from "./PickList.module.css";
import { SwipeRow } from "./SwipeRow";
import { TextField } from "./TextField";
import { useSheet } from "./useSheet";

type Category = {
  name: string;
  // The coach's own movements in it: the ones a delete leaves without a category.
  own: number;
  // A built-in movement uses it, so it can't be renamed or deleted.
  builtIn: boolean;
};

function categoriesOf(movements: Movement[]): Category[] {
  const byName = new Map<string, Category>();
  for (const m of movements) {
    if (!m.category) continue;
    const c = byName.get(m.category) ?? { name: m.category, own: 0, builtIn: false };
    if (m.coach_id === null) c.builtIn = true;
    else c.own++;
    byName.set(m.category, c);
  }
  return [...byName.values()].sort((a, b) => a.name.localeCompare(b.name, "fa"));
}

type CategoriesSheetProps = {
  open: boolean;
  onClose: () => void;
  // The whole library, which the categories are read from.
  movements: Movement[];
};

// Two steps in one sheet, like CopyPickerSheet: the categories in use, then one of them to rename.
export function CategoriesSheet({ open, onClose, movements }: CategoriesSheetProps) {
  // A new key each time the sheet opens, so it starts at the list: the sheet keeps its content
  // mounted while it slides down, and could reopen on the category it closed on.
  const [openings, setOpenings] = useState(0);
  const [wasOpen, setWasOpen] = useState(open);
  if (open !== wasOpen) {
    setWasOpen(open);
    if (open) setOpenings(openings + 1);
  }

  return (
    <BottomSheet open={open} onClose={onClose} title="دسته‌بندی‌ها">
      {open && <CategoryManager key={openings} movements={movements} />}
    </BottomSheet>
  );
}

function CategoryManager({ movements }: { movements: Movement[] }) {
  const queryClient = useQueryClient();
  const confirm = useSheet("delete-category");
  const [editing, setEditing] = useState<string | null>(null);
  const [swiped, setSwiped] = useState<string | null>(null);
  // Kept after the delete, so the confirmation keeps its text while it slides down.
  const [deleting, setDeleting] = useState<Category | null>(null);
  const categories = categoriesOf(movements);
  const current = categories.find((c) => c.name === editing && !c.builtIn);

  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.movements() });
  const remove = useMutation({
    mutationFn: (name: string) => deleteMovementCategory(name),
    onSuccess: async () => {
      await refresh();
      confirm.close();
      setEditing(null);
    },
    onError: (err) => refreshIfStale(err, refresh),
  });
  const askDelete = (c: Category) => {
    setSwiped(null);
    remove.reset();
    setDeleting(c);
    confirm.show();
  };

  return (
    <>
      {current ? (
        <CategoryForm
          key={current.name}
          category={current}
          categories={categories}
          onDelete={() => askDelete(current)}
          onDone={() => setEditing(null)}
        />
      ) : (
        <div className={s.scroll}>
          <div className={`${p.list} ${s.list}`}>
            {categories.length === 0 && <p className={p.note}>هنوز دسته‌ای نیست.</p>}
            {categories.map((c) =>
              // Built-in categories can't be renamed or deleted, so they don't swipe.
              c.builtIn ? (
                <div key={c.name} className={`${s.line} ${s.row}`}>
                  <span className={p.nm}>
                    <b>{c.name}</b>
                  </span>
                  <span className={s.uni}>
                    <LockIcon />
                    پیش‌فرض
                  </span>
                </div>
              ) : (
                <SwipeRow
                  key={c.name}
                  className={s.line}
                  open={swiped === c.name}
                  onOpenChange={(o) => setSwiped(o ? c.name : null)}
                  onEdit={() => {
                    setSwiped(null);
                    setEditing(c.name);
                  }}
                  onDelete={() => askDelete(c)}
                >
                  <button type="button" className={s.row} onClick={() => setEditing(c.name)}>
                    <span className={p.nm}>
                      <b>{c.name}</b>
                      <span>{toPersianDigits(c.own)} حرکت</span>
                    </span>
                    <span className={p.go}>
                      <ChevLIcon />
                    </span>
                  </button>
                </SwipeRow>
              ),
            )}
          </div>
        </div>
      )}
      {deleting && (
        <ConfirmSheet
          open={confirm.open}
          onClose={confirm.close}
          title={`حذف دسته «${deleting.name}»؟`}
          text={
            deleting.own > 0
              ? `${toPersianDigits(deleting.own)} حرکت بدون دسته می‌شوند.`
              : undefined
          }
          confirmLabel="حذف"
          onConfirm={() => remove.mutate(deleting.name)}
          pending={remove.isPending}
          error={remove.isError ? remove.error.message : null}
        />
      )}
    </>
  );
}

// A 403 or 404 means the list was stale: a built-in movement took the name, or it's already gone.
function refreshIfStale(err: Error, refresh: () => void) {
  if (
    err instanceof ApiError &&
    (err.code === ErrorCode.forbidden || err.code === ErrorCode.notFound)
  ) {
    refresh();
  }
}

type CategoryFormProps = {
  category: Category;
  categories: Category[];
  onDelete: () => void;
  onDone: () => void;
};

function CategoryForm({ category, categories, onDelete, onDone }: CategoryFormProps) {
  const queryClient = useQueryClient();
  const merge = useSheet("merge-category");
  const [name, setName] = useState(category.name);
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [nameError, setNameError] = useState<string | null>(null);
  const newName = normalizeCategory(name);

  // The builder's search and the movement sheet read the same query, so they see the change too.
  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.movements() });
  const rename = useMutation({
    mutationFn: () => renameMovementCategory(category.name, newName),
    onSuccess: async () => {
      await refresh();
      merge.close();
      onDone();
    },
    onError: (err) => refreshIfStale(err, refresh),
  });

  const errors = fieldErrors(rename.error);
  const submit = () => {
    if (!newName) {
      setNameError("نام دسته را وارد کنید.");
      rename.reset();
      return;
    }
    setNameError(null);
    if (newName === category.name) {
      onDone();
    } else if (categories.some((c) => c.name === newName)) {
      merge.show();
    } else {
      rename.mutate();
    }
  };

  return (
    <>
      <button type="button" className={s.back} onClick={onDone}>
        همه دسته‌ها
      </button>
      <TextField
        label="نام دسته"
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") submit();
        }}
        enterKeyHint="done"
        error={nameError ?? errors.name}
      />
      <div className={s.acts}>
        <button type="button" className={s.del} onClick={onDelete}>
          <TrashIcon /> حذف
        </button>
        <BigButton onClick={submit} disabled={rename.isPending}>
          ذخیره
        </BigButton>
      </div>
      {rename.isError && !merge.open && Object.keys(errors).length === 0 && (
        <p className={s.error} role="alert">
          {rename.error.message}
        </p>
      )}
      <ConfirmSheet
        open={merge.open}
        onClose={merge.close}
        title="یکی کردن دسته‌ها؟"
        text={`دسته «${category.name}» با «${newName}» یکی می‌شود.`}
        confirmLabel="یکی کن"
        onConfirm={() => rename.mutate()}
        pending={rename.isPending}
        error={rename.isError ? rename.error.message : null}
      />
    </>
  );
}
