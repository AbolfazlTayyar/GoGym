import { Link } from "react-router";
import type { Day } from "../api/types";
import { toPersianDigits } from "../lib/digits";
import s from "./PlanMini.module.css";

type PlanMiniProps = {
  days: Day[];
  // With a link per day, each box opens that day (the profile's current plan); without, it only shows.
  dayLink?: (label: string) => string;
};

export function PlanMini({ days, dayLink }: PlanMiniProps) {
  return (
    <div className={s.mini}>
      {days.map((d) => {
        const content = (
          <>
            <b>{d.label}</b>
            <span>{toPersianDigits(d.blocks.length)} حرکت</span>
          </>
        );
        return dayLink ? (
          <Link key={d.id} to={dayLink(d.label)} className={s.day}>
            {content}
          </Link>
        ) : (
          <div key={d.id}>{content}</div>
        );
      })}
    </div>
  );
}
