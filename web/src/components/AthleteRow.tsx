import type { CSSProperties, ReactNode } from "react";
import { Link } from "react-router";
import type { Athlete, PlanSummary } from "../api/types";
import { letterColumns, minDayColumns } from "../lib/days";
import { paths } from "../lib/paths";
import s from "./AthleteRow.module.css";
import { PlusIcon } from "./icons";

type AthleteRowProps = {
  athlete: Athlete;
  plan: PlanSummary | null;
  columns: number;
  // Search passes the query so the matching part of the name is marked.
  highlight?: string;
};

export function AthleteRow({ athlete, plan, columns, highlight }: AthleteRowProps) {
  const name = `${athlete.first_name} ${athlete.last_name}`;
  // The name opens the current plan on its first day, or the profile when there is none.
  const nameTo = plan ? paths.dayView(plan.id) : paths.profile(athlete.id);

  return (
    <div className={s.ath}>
      <Link to={nameTo} className={s.nm}>
        <b>{highlight ? highlighted(name, highlight) : name}</b>
        {plan ? <span>{plan.title}</span> : <span className={s.none}>بدون برنامه</span>}
      </Link>
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
