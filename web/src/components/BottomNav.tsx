import { NavLink } from "react-router";
import s from "./BottomNav.module.css";
import { DumbbellIcon, UsersIcon } from "./icons";

const itemClass = ({ isActive }: { isActive: boolean }) =>
  isActive ? `${s.item} ${s.on}` : s.item;

// The app's only two top-level pages (docs/design/README.md, "How they connect").
export function BottomNav() {
  return (
    <nav className={s.nav}>
      <NavLink to="/" end className={itemClass}>
        <UsersIcon /> شاگردها
      </NavLink>
      <NavLink to="/movements" className={itemClass}>
        <DumbbellIcon /> حرکات
      </NavLink>
    </nav>
  );
}
