import type { ReactNode } from "react";
import s from "./EmptyState.module.css";

type EmptyStateProps = {
  icon?: ReactNode;
  title: ReactNode;
  text?: ReactNode;
  action?: ReactNode;
};

export function EmptyState({ icon, title, text, action }: EmptyStateProps) {
  return (
    <div className={s.empty}>
      {icon && <div className={s.icon}>{icon}</div>}
      <div className={s.title}>{title}</div>
      {text && <div className={s.text}>{text}</div>}
      {action && <div className={s.action}>{action}</div>}
    </div>
  );
}
