import { useLocation, useNavigate } from "react-router";

type SheetState = { sheets?: string[] } | null;

// Which sheets are open lives in the history entry's state, so opening one pushes an entry and the
// phone's back button pops it. A stack, so a date picker can open over a form sheet.
export function useSheet(name: string) {
  const location = useLocation();
  const navigate = useNavigate();
  const state = location.state as SheetState;
  const sheets = state?.sheets ?? [];
  const open = sheets.includes(name);

  const show = () => {
    if (open) return;
    navigate(
      { pathname: location.pathname, search: location.search, hash: location.hash },
      { state: { ...state, sheets: [...sheets, name] } },
    );
  };

  const close = () => {
    if (open) navigate(-1);
  };

  return { open, show, close };
}
