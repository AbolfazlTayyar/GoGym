import { useState } from "react";
import { useLocation, useNavigate } from "react-router";

// Back to wherever the coach came from. A page opened directly, with no app history behind it, goes to
// fallback instead. entriesAbove counts history entries pushed on top of the page since, such as an open
// sheet's, so leaving from a sheet takes the page with it in one step.
export function useGoBack(fallback: string) {
  const navigate = useNavigate();
  const { key } = useLocation();
  const [openedDirectly] = useState(key === "default");

  return (entriesAbove = 0) => {
    if (openedDirectly) navigate(fallback, { replace: true });
    else navigate(-1 - entriesAbove);
  };
}
