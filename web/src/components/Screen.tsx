import type { ReactNode } from "react";
import s from "./Screen.module.css";

export function Screen({ children }: { children: ReactNode }) {
  return <div className={s.screen}>{children}</div>;
}

export function ScreenContent({ children }: { children: ReactNode }) {
  return <main className={s.content}>{children}</main>;
}
