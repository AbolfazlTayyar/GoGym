import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { listAthletes } from "../api/athletes";
import { queryKeys } from "../api/keys";
import { listAthletePlans } from "../api/plans";
import type { Athlete, PlanSummary } from "../api/types";
import { formatJalali } from "../lib/jalali";
import { athleteTypeLabels } from "../lib/labels";
import { planSizeText } from "../lib/planSize";
import { useDebouncedValue } from "../lib/useDebouncedValue";
import { BottomSheet } from "./BottomSheet";
import s from "./CopyPickerSheet.module.css";
import { ChevLIcon } from "./icons";
import p from "./PickList.module.css";
import { SearchBox } from "./SearchBox";

const searchDebounceMs = 250;

type CopyPickerSheetProps = {
  open: boolean;
  onClose: () => void;
  // The athlete the new plan is for, left out of the list.
  exceptAthleteId: string;
  onPick: (athlete: Athlete, plan: PlanSummary) => void;
};

// Two steps in one sheet: find the athlete with the same search as home, then pick one of their plans.
export function CopyPickerSheet({ open, onClose, exceptAthleteId, onPick }: CopyPickerSheetProps) {
  const [text, setText] = useState("");
  const q = useDebouncedValue(text.trim(), searchDebounceMs);
  const [athlete, setAthlete] = useState<Athlete | null>(null);

  const athletes = useQuery({
    queryKey: queryKeys.athletes("search", q),
    queryFn: () => listAthletes({ type: "all", q }),
    enabled: open && athlete === null,
  });
  const plans = useQuery({
    queryKey: queryKeys.plans("athlete", athlete?.id),
    queryFn: () => listAthletePlans(athlete!.id),
    enabled: athlete !== null,
  });

  const others = (athletes.data?.data ?? []).filter((a) => a.id !== exceptAthleteId);

  return (
    <BottomSheet open={open} onClose={onClose} title="کپی برنامه یک شاگرد دیگر">
      {athlete === null ? (
        <>
          <div className={s.search}>
            <SearchBox value={text} onChange={setText} placeholder="جستجوی نام شاگرد…" />
          </div>
          <div className={s.scroll}>
            <div className={p.list}>
              {athletes.isPending && <p className={p.note}>در حال بارگذاری…</p>}
              {athletes.isError && <p className={p.note}>{athletes.error.message}</p>}
              {athletes.isSuccess && others.length === 0 && <p className={p.note}>کسی پیدا نشد.</p>}
              {others.map((a) => (
                <button key={a.id} type="button" className={p.row} onClick={() => setAthlete(a)}>
                  <span className={p.nm}>
                    <b>
                      {a.first_name} {a.last_name}
                    </b>
                    <span>{athleteTypeLabels[a.athlete_type]}</span>
                  </span>
                  <span className={p.go}>
                    <ChevLIcon />
                  </span>
                </button>
              ))}
            </div>
          </div>
        </>
      ) : (
        <>
          <button type="button" className={s.back} onClick={() => setAthlete(null)}>
            {athlete.first_name} {athlete.last_name} · شاگرد دیگر
          </button>
          <div className={s.scroll}>
            <div className={p.list}>
              {plans.isPending && <p className={p.note}>در حال بارگذاری…</p>}
              {plans.isError && <p className={p.note}>{plans.error.message}</p>}
              {plans.isSuccess && plans.data.length === 0 && (
                <p className={p.note}>این شاگرد هنوز برنامه‌ای ندارد.</p>
              )}
              {plans.data?.map((plan) => (
                <button
                  key={plan.id}
                  type="button"
                  className={p.row}
                  onClick={() => onPick(athlete, plan)}
                >
                  <span className={p.nm}>
                    <b>{plan.title}</b>
                    <span>
                      شروع {formatJalali(plan.start_date)} · {planSizeText(plan)}
                    </span>
                  </span>
                  <span className={p.go}>
                    <ChevLIcon />
                  </span>
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </BottomSheet>
  );
}
