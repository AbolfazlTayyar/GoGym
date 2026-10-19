import type { ReactNode } from "react";
import { Link } from "react-router";
import s from "./BigButton.module.css";

type BigButtonProps = {
  children: ReactNode;
  type?: "button" | "submit";
  onClick?: () => void;
  disabled?: boolean;
  to?: string;
  outline?: boolean;
};

export function BigButton({
  children,
  type = "button",
  onClick,
  disabled,
  to,
  outline,
}: BigButtonProps) {
  const className = outline ? `${s.big} ${s.outline}` : s.big;
  if (to !== undefined) {
    return (
      <Link to={to} className={className}>
        {children}
      </Link>
    );
  }
  return (
    <button type={type} className={className} onClick={onClick} disabled={disabled}>
      {children}
    </button>
  );
}

export function BottomBar({ children }: { children: ReactNode }) {
  return <div className={s.bottom}>{children}</div>;
}
