import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router";
import { listAthletes } from "../api/athletes";
import { queryKeys } from "../api/keys";
import { getCurrentPlans } from "../api/plans";
import type { Athlete, AthleteType, PlanSummary } from "../api/types";
import { AthleteRow } from "../components/AthleteRow";
import { BigButton } from "../components/BigButton";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, KickerButton, RoundButton } from "../components/Header";
import { PlusIcon, SparkIcon, UsersIcon } from "../components/icons";
import { PillTabs } from "../components/PillTabs";
import { SearchBox } from "../components/SearchBox";
import { SignOutSheet } from "../components/SignOutSheet";
import { useSheet } from "../components/useSheet";
import { dayColumns } from "../lib/days";
import { toPersianDigits } from "../lib/digits";
import { paths } from "../lib/paths";
import { useDebouncedValue } from "../lib/useDebouncedValue";
import { useScrollMemory } from "../lib/useScrollMemory";
import s from "./HomePage.module.css";

const typeParam = "type";
const searchParam = "q";
const searchDebounceMs = 250;

const typeLabels: Record<AthleteType, string> = { private: "خصوصی", public: "عمومی" };

export function HomePage() {
  const [params, setParams] = useSearchParams();
  const type: AthleteType = params.get(typeParam) === "public" ? "public" : "private";
  const [text, setText] = useState(params.get(searchParam) ?? "");
  const q = useDebouncedValue(text.trim(), searchDebounceMs);
  const account = useSheet("account");
  const rosterRef = useRef<HTMLDivElement>(null);

  // The search lives in the URL too, so back from a day opened from the results returns to them.
  useEffect(() => {
    if ((params.get(searchParam) ?? "") === q) return;
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (q) next.set(searchParam, q);
        else next.delete(searchParam);
        return next;
      },
      { replace: true },
    );
  }, [q, params, setParams]);

  // Both tabs come from one request; with 20-30 athletes every one fits in it.
  const athletes = useQuery({
    queryKey: queryKeys.athletes("all"),
    queryFn: () => listAthletes({ type: "all" }),
  });
  const current = useQuery({ queryKey: queryKeys.plans("current"), queryFn: getCurrentPlans });
  // The API does the matching, across private and public together.
  const search = useQuery({
    queryKey: queryKeys.athletes("search", q),
    queryFn: () => listAthletes({ type: "all", q }),
    enabled: q !== "",
    placeholderData: keepPreviousData,
  });

  const searching = text.trim() !== "";
  const ready = athletes.isSuccess && current.isSuccess && (!searching || search.isSuccess);
  useScrollMemory(rosterRef, ready);

  const plans = new Map<string, PlanSummary>((current.data ?? []).map((p) => [p.athlete_id, p]));
  const columns = dayColumns((current.data ?? []).map((p) => p.day_labels));
  const all = athletes.data?.data ?? [];
  const ofType = (list: Athlete[], t: AthleteType) => list.filter((a) => a.athlete_type === t);

  const rows = (list: Athlete[], highlight?: string) =>
    list.map((a) => (
      <AthleteRow
        key={a.id}
        athlete={a}
        plan={plans.get(a.id) ?? null}
        columns={columns}
        highlight={highlight}
      />
    ));

  const setType = (next: AthleteType) =>
    setParams(
      (prev) => {
        const p = new URLSearchParams(prev);
        if (next === "public") p.set(typeParam, next);
        else p.delete(typeParam);
        return p;
      },
      { replace: true },
    );

  const failed = athletes.error ?? current.error ?? (searching ? search.error : null);

  let body;
  if (failed) {
    body = (
      <div className={s.fill}>
        <EmptyState
          title={failed.message}
          action={
            <BigButton
              onClick={() => {
                athletes.refetch();
                current.refetch();
                if (searching) search.refetch();
              }}
            >
              دوباره
            </BigButton>
          }
        />
      </div>
    );
  } else if (!ready) {
    body = (
      <div className={s.roster} aria-busy>
        <p className={s.note}>در حال بارگذاری…</p>
      </div>
    );
  } else if (searching) {
    const found = search.data?.data ?? [];
    body = (
      <div ref={rosterRef} className={`${s.roster} ${s.res}`}>
        {found.length === 0 && <p className={s.note}>کسی با این نام پیدا نشد.</p>}
        {(["private", "public"] as const).map((t) => {
          const group = ofType(found, t);
          return (
            group.length > 0 && (
              <section key={t}>
                <div className={s.gl}>
                  {typeLabels[t]} · {toPersianDigits(group.length)}
                </div>
                {rows(group, q)}
              </section>
            )
          );
        })}
      </div>
    );
  } else if (ofType(all, type).length === 0) {
    body = (
      <div className={s.fill}>
        <EmptyState
          icon={<UsersIcon />}
          title={all.length === 0 ? "هنوز شاگردی ندارید" : `شاگرد ${typeLabels[type]} ندارید`}
          action={
            <BigButton to={paths.newAthlete(type)}>+ شاگرد {typeLabels[type]} جدید</BigButton>
          }
        />
      </div>
    );
  } else {
    body = (
      <div ref={rosterRef} className={s.roster}>
        {rows(ofType(all, type))}
      </div>
    );
  }

  return (
    <>
      <Header>
        <HeaderTitle
          kicker={
            <KickerButton onClick={account.show}>
              <SparkIcon /> سلام مربی
            </KickerButton>
          }
          title="شاگردهای من"
        />
        <RoundButton label="شاگرد جدید" pink to={paths.newAthlete(type)}>
          <PlusIcon />
        </RoundButton>
      </Header>
      <SearchBox value={text} onChange={setText} placeholder="جستجوی نام شاگرد…" />
      {!searching && (
        <PillTabs
          options={(["private", "public"] as const).map((t) => ({
            value: t,
            label: `${typeLabels[t]} ${toPersianDigits(ofType(all, t).length)}`,
          }))}
          value={type}
          onChange={setType}
        />
      )}
      {body}
      <SignOutSheet open={account.open} onClose={account.close} />
    </>
  );
}
