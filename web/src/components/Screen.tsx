import type { CSSProperties, ReactNode } from "react";
import s from "./Screen.module.css";

export function Screen({ children, style }: { children: ReactNode; style?: CSSProperties }) {
  return (
    <div className={s.screen} style={style}>
      {children}
    </div>
  );
}

export function ScreenContent({ children }: { children: ReactNode }) {
  return <main className={s.content}>{children}</main>;
}
