import type { ReactNode } from "react";
import { Link } from "react-router";
import s from "./Header.module.css";

export function Header({ children }: { children: ReactNode }) {
  return <header className={s.hdr}>{children}</header>;
}

type HeaderTitleProps = {
  title: ReactNode;
  subtitle?: ReactNode;
  // The small pink line above a top-level page's large title.
  kicker?: ReactNode;
};

// Top-level pages (with a kicker) get the board's large title; inner pages the smaller one with a subtitle.
export function HeaderTitle({ title, subtitle, kicker }: HeaderTitleProps) {
  return (
    <div className={s.t}>
      {kicker !== undefined && <div className={s.hi}>{kicker}</div>}
      {kicker !== undefined ? <h1 className={s.h3}>{title}</h1> : <h1 className={s.b}>{title}</h1>}
      {subtitle !== undefined && <span className={s.s}>{subtitle}</span>}
    </div>
  );
}

// A kicker the coach can tap, such as the greeting that opens the sign-out sheet.
export function KickerButton({ children, onClick }: { children: ReactNode; onClick: () => void }) {
  return (
    <button type="button" className={s.kickerButton} onClick={onClick}>
      {children}
    </button>
  );
}

type RoundButtonProps = {
  label: string;
  children: ReactNode;
  pink?: boolean;
  to?: string;
  onClick?: () => void;
};

export function RoundButton({ label, children, pink, to, onClick }: RoundButtonProps) {
  const className = pink ? `${s.rb} ${s.pink}` : s.rb;
  if (to !== undefined) {
    return (
      <Link to={to} className={className} aria-label={label}>
        {children}
      </Link>
    );
  }
  return (
    <button type="button" className={className} aria-label={label} onClick={onClick}>
      {children}
    </button>
  );
}

type PillButtonProps = {
  children: ReactNode;
  onClick?: () => void;
  disabled?: boolean;
};

export function PillButton({ children, onClick, disabled }: PillButtonProps) {
  return (
    <button type="button" className={s.pbtn} onClick={onClick} disabled={disabled}>
      {children}
    </button>
  );
}
