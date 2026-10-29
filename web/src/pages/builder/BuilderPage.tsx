import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useReducer, useRef, useState } from "react";
import { useBlocker, useParams, useSearchParams } from "react-router";
import { getAthlete } from "../../api/athletes";
import { ApiError } from "../../api/client";
import { queryKeys } from "../../api/keys";
import { addDay, deleteDay, getPlan, saveDay } from "../../api/plans";
import type { Athlete, Day, PlanDetail } from "../../api/types";
import { BigButton } from "../../components/BigButton";
import { BlockLine, DayHeading } from "../../components/BlockLine";
import { ConfirmSheet } from "../../components/ConfirmSheet";
import { DayPills } from "../../components/DayPills";
import pills from "../../components/DayPills.module.css";
import { EmptyState } from "../../components/EmptyState";
import { Header, HeaderTitle, PillButton, RoundButton } from "../../components/Header";
import { BackIcon, PlusIcon, TrashIcon } from "../../components/icons";
import { Screen } from "../../components/Screen";
import { useSheet } from "../../components/useSheet";
import { paths } from "../../lib/paths";
import { useGoBack } from "../../lib/useGoBack";
import { useVisualViewportStyle } from "../../lib/useVisualViewport";
import { Adder } from "./Adder";
import s from "./Builder.module.css";
import { draftReducer, fromDay, newKey, toBody } from "./draft";
import { OpenLine } from "./OpenLine";

const dayParam = "day";
const dayLetters = "ABCDEFG";
const savedNoticeMs = 2000;

// The next day "+" adds: the first slot whose letter and order_index are both free.
function nextDay(days: Day[]): { label: string; orderIndex: number } | null {
  for (let i = 0; i < dayLetters.length; i++) {
    if (!days.some((d) => d.order_index === i || d.label === dayLetters[i])) {
      return { label: dayLetters[i], orderIndex: i };
    }
  }
  return null;
}

// The API keys a save's errors by position, "[2].movements[1].movement_id"; this groups them by line.
function lineErrors(err: unknown): Map<number, string> {
  const byLine = new Map<number, string>();
  if (!(err instanceof ApiError)) return byLine;
  for (const [key, message] of Object.entries(err.fields)) {
    const match = /^\[(\d+)\]/.exec(key);
    if (!match) continue;
    const i = Number(match[1]);
    byLine.set(i, byLine.has(i) ? `${byLine.get(i)} ${message}` : message);
  }
  return byLine;
}

export function BuilderPage() {
  const { planId = "" } = useParams();
  const [params] = useSearchParams();
  const plan = useQuery({
    queryKey: queryKeys.plans("detail", planId),
    queryFn: () => getPlan(planId),
  });
  const athleteId = plan.data?.athlete_id ?? "";
  const athlete = useQuery({
    queryKey: queryKeys.athletes("detail", athleteId),
    queryFn: () => getAthlete(athleteId),
    enabled: athleteId !== "",
  });

  const failed = plan.error ?? athlete.error;
  if (failed) {
    return (
      <Screen>
        <EmptyState
          title={failed.message}
          action={<BigButton to={paths.home}>بازگشت به شاگردها</BigButton>}
        />
      </Screen>
    );
  }
  if (!plan.data || !athlete.data) return <Screen>{null}</Screen>;

  const day = plan.data.days.find((d) => d.label === params.get(dayParam)) ?? plan.data.days[0];
  // Keyed by day: switching days starts a fresh draft from that day's saved lines.
  return (
    <BuilderScreen key={day?.id ?? "no-day"} plan={plan.data} athlete={athlete.data} day={day} />
  );
}

type BuilderScreenProps = { plan: PlanDetail; athlete: Athlete; day: Day | undefined };

function BuilderScreen({ plan, athlete, day }: BuilderScreenProps) {
  const queryClient = useQueryClient();
  const [, setParams] = useSearchParams();
  const goBack = useGoBack(paths.dayView(plan.id, day?.label));
  const viewport = useVisualViewportStyle();
  const deleteDaySheet = useSheet("delete-day");

  const [baseline, setBaseline] = useState(day);
  const [draft, dispatch] = useReducer(draftReducer, day, (d) => (d ? fromDay(d) : []));
  const [openKey, setOpenKey] = useState<string | null>(null);
  const [joinKey, setJoinKey] = useState<string | null>(null);
  const [pending, setPending] = useState<{ label: string } | "add" | null>(null);
  const [savedNotice, setSavedNotice] = useState(false);
  const sheetRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  const dirty =
    baseline !== undefined &&
    JSON.stringify(toBody(draft)) !== JSON.stringify(toBody(fromDay(baseline)));

  const detailKey = queryKeys.plans("detail", plan.id);
  const setDays = (days: (prev: Day[]) => Day[]) =>
    queryClient.setQueryData<PlanDetail>(detailKey, (p) => p && { ...p, days: days(p.days) });
  // Summaries (sizes, letters) on home and the profile follow the plan's days.
  const refreshSummaries = () =>
    queryClient.invalidateQueries({
      predicate: (q) => q.queryKey[0] === "plans" && q.queryKey[1] !== "detail",
    });

  const save = useMutation({
    mutationFn: () => saveDay(day!.id, toBody(draft)),
    onSuccess: (saved) => {
      setBaseline(saved);
      dispatch({ type: "load", day: saved });
      setOpenKey(null);
      setJoinKey(null);
      setDays((days) => days.map((d) => (d.id === saved.id ? saved : d)));
      refreshSummaries();
      setSavedNotice(true);
    },
  });
  const errors = lineErrors(save.error);
  const saveFailed = save.isError && errors.size === 0;

  useEffect(() => {
    if (!savedNotice) return;
    const timer = setTimeout(() => setSavedNotice(false), savedNoticeMs);
    return () => clearTimeout(timer);
  }, [savedNotice]);

  const add = useMutation({
    mutationFn: (next: { label: string; orderIndex: number }) =>
      addDay(plan.id, next.label, next.orderIndex),
    onSuccess: (created) => {
      setDays((days) => [...days, created].sort((a, b) => a.order_index - b.order_index));
      refreshSummaries();
      setParams({ [dayParam]: created.label }, { replace: true });
    },
  });

  const removeDay = useMutation({
    mutationFn: () => deleteDay(day!.id),
    onSuccess: () => {
      const rest = plan.days.filter((d) => d.id !== day!.id);
      setDays(() => rest);
      refreshSummaries();
      // Replacing the sheet's entry: back mustn't return to a day that is gone.
      setParams(rest[0] ? { [dayParam]: rest[0].label } : {}, { replace: true });
    },
  });

  // Back, the bottom bar or any link out of the builder asks first while the draft has changes.
  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) =>
      dirty && currentLocation.pathname !== nextLocation.pathname,
  );
  // Closing the tab or reloading too; the browser shows its own question.
  useEffect(() => {
    if (!dirty) return;
    const warn = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  // A new line opens and scrolls into view, so its reps can be set right away.
  const lastKey = draft.at(-1)?.key;
  useEffect(() => {
    if (lastKey && lastKey === openKey) {
      sheetRef.current?.scrollTo({ top: sheetRef.current.scrollHeight, behavior: "smooth" });
    }
  }, [lastKey, openKey]);

  const switchDay = (label: string) => {
    if (label === day?.label) return;
    if (dirty) setPending({ label });
    else setParams({ [dayParam]: label }, { replace: true });
  };
  const next = nextDay(plan.days);
  const addNextDay = () => {
    if (!next) return;
    if (dirty) setPending("add");
    else add.mutate(next);
  };
  const discardAndGo = () => {
    const target = pending;
    setPending(null);
    if (target === "add") {
      if (next) add.mutate(next);
    } else if (target) {
      setParams({ [dayParam]: target.label }, { replace: true });
    }
  };

  const onPick = (movement: { id: string; name: string }) => {
    if (joinKey) {
      dispatch({ type: "join", blockKey: joinKey, movementKey: newKey(), movement });
      setOpenKey(joinKey);
      setJoinKey(null);
      return;
    }
    const blockKey = newKey();
    dispatch({ type: "add", blockKey, movementKey: newKey(), movement });
    setOpenKey(blockKey);
  };

  const subtitle = `${athlete.first_name} ${athlete.last_name}${
    plan.source_plan_title ? ` · کپی از ${plan.source_plan_title}` : ""
  }`;

  return (
    <Screen style={viewport}>
      <Header>
        <RoundButton label="بازگشت" onClick={() => goBack()}>
          <BackIcon />
        </RoundButton>
        <HeaderTitle title={plan.title} subtitle={subtitle} />
        <PillButton onClick={() => save.mutate()} disabled={!day || save.isPending}>
          {savedNotice ? "ذخیره شد ✓" : "ذخیره"}
        </PillButton>
      </Header>
      <DayPills
        labels={plan.days.map((d) => d.label)}
        current={day?.label ?? null}
        onSelect={switchDay}
        trailing={
          next && (
            <button
              type="button"
              className={pills.add}
              aria-label={`افزودن روز ${next.label}`}
              onClick={addNextDay}
              disabled={add.isPending}
            >
              <PlusIcon />
            </button>
          )
        }
      />
      <div ref={sheetRef} className={s.sheet}>
        {!day ? (
          <p className={s.note}>این برنامه هنوز روزی ندارد؛ با + یک روز بسازید.</p>
        ) : (
          <>
            <div className={s.dhRow}>
              <DayHeading label={day.label} lines={draft.length} />
              <button
                type="button"
                className={s.dayDel}
                aria-label={`حذف روز ${day.label}`}
                onClick={deleteDaySheet.show}
              >
                <TrashIcon />
              </button>
            </div>
            {saveFailed && (
              <p className={s.banner} role="alert">
                ذخیره نشد: {save.error.message} تغییرات شما هنوز اینجاست.
              </p>
            )}
            {draft.length === 0 && (
              <p className={s.note}>حرکت را پایین جستجو کنید تا به این روز اضافه شود.</p>
            )}
            {draft.map((block, i) =>
              block.key === openKey ? (
                <OpenLine
                  key={block.key}
                  index={i}
                  block={block}
                  isFirst={i === 0}
                  isLast={i === draft.length - 1}
                  joining={joinKey === block.key}
                  dispatch={dispatch}
                  onClose={() => setOpenKey(null)}
                  onJoin={() => {
                    setJoinKey(joinKey === block.key ? null : block.key);
                    searchRef.current?.focus();
                  }}
                  error={errors.get(i)}
                />
              ) : (
                <div
                  key={block.key}
                  role="button"
                  tabIndex={0}
                  className={s.line}
                  onClick={() => setOpenKey(block.key)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") setOpenKey(block.key);
                  }}
                >
                  <BlockLine index={i} block={block} />
                  {errors.has(i) && <span className={s.lineError}>{errors.get(i)}</span>}
                </div>
              ),
            )}
          </>
        )}
      </div>
      {day && <Adder inputRef={searchRef} joining={joinKey !== null} onPick={onPick} />}

      {day && (
        <ConfirmSheet
          open={deleteDaySheet.open}
          onClose={deleteDaySheet.close}
          title={`حذف روز ${day.label}؟`}
          text="همه حرکت‌های این روز پاک می‌شوند. روزهای دیگر سر جایشان می‌مانند."
          confirmLabel="حذف روز"
          onConfirm={() => removeDay.mutate()}
          pending={removeDay.isPending}
          error={removeDay.isError ? removeDay.error.message : null}
        />
      )}
      <ConfirmSheet
        open={pending !== null || blocker.state === "blocked"}
        onClose={() => (blocker.state === "blocked" ? blocker.reset() : setPending(null))}
        title={`تغییرات روز ${day?.label ?? ""} ذخیره نشده‌اند`}
        text="اگر ادامه دهید، تغییراتی که دادید از بین می‌روند. برای نگه داشتنشان اول ذخیره کنید."
        confirmLabel="ادامه بدون ذخیره"
        onConfirm={() => (blocker.state === "blocked" ? blocker.proceed() : discardAndGo())}
      />
    </Screen>
  );
}
