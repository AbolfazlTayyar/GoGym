import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { queryKeys } from "../api/keys";
import { listMovements } from "../api/movements";
import type { Movement } from "../api/types";
import { BigButton } from "../components/BigButton";
import { ChipRow } from "../components/ChipRow";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { ChevLIcon, DumbbellIcon, LockIcon, PlusIcon } from "../components/icons";
import { MovementSheet } from "../components/MovementSheet";
import { SearchBox } from "../components/SearchBox";
import { useSheet } from "../components/useSheet";
import { toPersianDigits } from "../lib/digits";
import { useDebouncedValue } from "../lib/useDebouncedValue";
import { useScrollMemory } from "../lib/useScrollMemory";
import s from "./MovementsPage.module.css";

const searchDebounceMs = 250;
// The "همه" chip's value; categories are free text, so it can't be one of them.
const allCategories = "\u0000all";

export function MovementsPage() {
  const [text, setText] = useState("");
  const q = useDebouncedValue(text.trim(), searchDebounceMs);
  const [category, setCategory] = useState(allCategories);
  const sheet = useSheet("movement");
  const [selected, setSelected] = useState<Movement | null>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // The same query the builder's search uses.
  const movements = useQuery({
    queryKey: queryKeys.movements("list", q),
    queryFn: () => listMovements(q),
    placeholderData: keepPreviousData,
  });
  // The counts and the category chips describe the whole library, not the search results.
  const library = useQuery({
    queryKey: queryKeys.movements("list", ""),
    queryFn: () => listMovements(),
  });
  useScrollMemory(listRef, movements.isSuccess);

  const all = library.data ?? [];
  const own = all.filter((m) => m.coach_id !== null).length;
  const categories = [...new Set(all.flatMap((m) => (m.category ? [m.category] : [])))].sort(
    (a, b) => a.localeCompare(b, "fa"),
  );
  const shown = (movements.data ?? []).filter(
    (m) => category === allCategories || m.category === category,
  );

  const open = (m: Movement | null) => {
    setSelected(m);
    sheet.show();
  };

  let body;
  if (movements.isError) {
    body = (
      <div className={s.fill}>
        <EmptyState
          title={movements.error.message}
          action={<BigButton onClick={() => movements.refetch()}>دوباره</BigButton>}
        />
      </div>
    );
  } else {
    body = (
      <div ref={listRef} className={s.roster}>
        {movements.isPending && <p className={s.note}>در حال بارگذاری…</p>}
        {movements.isSuccess && shown.length === 0 && <p className={s.note}>حرکتی پیدا نشد.</p>}
        {shown.map((m) => (
          <button key={m.id} type="button" className={s.mvr} onClick={() => open(m)}>
            <div className={s.nm}>
              <b>{m.name}</b>
              <span>{m.category ?? "بدون دسته"}</span>
            </div>
            {m.coach_id === null ? (
              <span className={s.uni}>
                <LockIcon />
                پیش‌فرض
              </span>
            ) : (
              <span className={s.go}>
                <ChevLIcon />
              </span>
            )}
          </button>
        ))}
      </div>
    );
  }

  return (
    <>
      <Header>
        <HeaderTitle
          kicker={`${toPersianDigits(own)} حرکت شما · ${toPersianDigits(all.length - own)} پیش‌فرض`}
          title="حرکات"
        />
        <RoundButton label="حرکت جدید" pink onClick={() => open(null)}>
          <PlusIcon />
        </RoundButton>
      </Header>
      <SearchBox value={text} onChange={setText} placeholder="جستجوی حرکت…" />
      <div className={s.chips}>
        <ChipRow
          options={[
            { value: allCategories, label: `همه ${toPersianDigits(all.length)}` },
            ...categories.map((c) => ({ value: c, label: c })),
          ]}
          value={category}
          onChange={setCategory}
        />
      </div>
      {library.isSuccess && all.length === 0 ? (
        <div className={s.fill}>
          <EmptyState icon={<DumbbellIcon />} title="هنوز حرکتی نیست" />
        </div>
      ) : (
        body
      )}
      <MovementSheet
        open={sheet.open}
        onClose={sheet.close}
        movement={selected}
        categories={categories}
      />
    </>
  );
}
