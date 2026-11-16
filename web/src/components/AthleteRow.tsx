import type { CSSProperties, ReactNode } from "react";
import { Link } from "react-router";
import type { Athlete, AttendanceCount, PlanSummary } from "../api/types";
import { msgSessionsDone, sessionsUsed, visitCountText } from "../lib/attendance";
import { letterColumns, minDayColumns } from "../lib/days";
import { todayISO } from "../lib/jalali";
import { paths } from "../lib/paths";
import s from "./AthleteRow.module.css";
import { PlusIcon } from "./icons";

type AthleteRowProps = {
  athlete: Athlete;
  plan: PlanSummary | null;
  columns: number;
  // Search passes the query so the matching part of the name is marked.
  highlight?: string;
  // Set when the current plan counts sessions: its progress, and the tap that records today's visit.
  sessions?: SessionsProps;
};

type SessionsProps = { count: AttendanceCount; onVisit: () => void; pending: boolean };

export function AthleteRow({ athlete, plan, columns, highlight, sessions }: AthleteRowProps) {
  const name = `${athlete.first_name} ${athlete.last_name}`;
  // The name opens the current plan on its first day, or the profile when there is none.
  const nameTo = plan ? paths.dayView(plan.id) : paths.profile(athlete.id);

  return (
    <div className={s.ath}>
      {/* The link stretches over the whole name column, so the sessions button can sit inside it. */}
      <div className={s.nm}>
        <Link to={nameTo} className={s.stretch}>
          <b>{highlight ? highlighted(name, highlight) : name}</b>
        </Link>
        <div className={s.sub}>
          {sessions && <Sessions {...sessions} />}
          {plan ? <span>{plan.title}</span> : <span className={s.none}>بدون برنامه</span>}
        </div>
      </div>
      <div
        className={columns > minDayColumns ? `${s.dl} ${s.dense}` : s.dl}
        style={{ "--columns": columns } as CSSProperties}
      >
        {plan ? (
          letterColumns(plan.day_labels, columns).map((label, i) =>
            label ? (
              <Link key={label} to={paths.dayView(plan.id, label)} className={s.letter}>
                {label}
              </Link>
            ) : (
              <span key={`empty-${i}`} data-testid="empty-day" />
            ),
          )
        ) : (
          <Link to={paths.newPlan(athlete.id)} className={s.new}>
            <PlusIcon />
            برنامه جدید
          </Link>
        )}
      </div>
    </div>
  );
}

// The 12th visit turns the button into a badge that stays until the next plan starts.
function Sessions({ count, onVisit, pending }: SessionsProps) {
  const text = visitCountText(count.visits, count.sessions);
  if (sessionsUsed(count.visits, count.sessions)) {
    return <b className={`${s.chip} ${s.done}`}>{msgSessionsDone}</b>;
  }
  // One visit a day: once today is in, the button only shows the count.
  if (count.last_date === todayISO()) {
    return (
      <b className={`${s.chip} ${s.today}`} aria-label={`${text} جلسه، امروز ثبت شده`}>
        ✓ {text}
      </b>
    );
  }
  return (
    <button
      type="button"
      className={s.chip}
      onClick={onVisit}
      disabled={pending}
      aria-label={`ثبت حضور امروز، ${text} جلسه`}
    >
      + {text}
    </button>
  );
}

// Marks every case-insensitive occurrence of query in text; the API already decided the row matches.
function highlighted(text: string, query: string): ReactNode[] {
  const parts: ReactNode[] = [];
  const lower = text.toLowerCase();
  const needle = query.trim().toLowerCase();
  let from = 0;
  for (let at = lower.indexOf(needle); needle && at >= 0; at = lower.indexOf(needle, from)) {
    parts.push(text.slice(from, at), <em key={at}>{text.slice(at, at + needle.length)}</em>);
    from = at + needle.length;
  }
  parts.push(text.slice(from));
  return parts;
}
