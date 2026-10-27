import type { KeyboardEvent } from "react";

// Phone keyboards have no Tab, so a form's onKeyDown lets Enter in an input move to the next text field;
// in the last one Enter keeps its default (submit), and in a textarea it still starts a new line.
export function focusNextOnEnter(e: KeyboardEvent<HTMLElement>) {
  if (e.key !== "Enter" || !(e.target instanceof HTMLInputElement)) return;
  const fields = [...e.currentTarget.querySelectorAll<HTMLElement>("input, textarea")];
  const next = fields[fields.indexOf(e.target) + 1];
  if (!next) return;
  e.preventDefault();
  next.focus();
}
