import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, type RefObject } from "react";
import { queryKeys } from "../../api/keys";
import { createMovement, listMovements } from "../../api/movements";
import type { Movement } from "../../api/types";
import { SearchIcon } from "../../components/icons";
import { useDebouncedValue } from "../../lib/useDebouncedValue";
import s from "./Builder.module.css";

// Longer than the gap between keys on a phone, so the chips change when the coach pauses, not at each character.
const searchDebounceMs = 400;
const maxSuggestions = 8;

type AdderProps = {
  inputRef: RefObject<HTMLInputElement | null>;
  // Set while "+ سوپرست" waits for the movement that joins a line.
  joining: boolean;
  onPick: (movement: { id: string; name: string }) => void;
};

// The search pinned under the day: the only typing in the builder. A pick adds a line; a name that isn't
// in the library yet is added to it first.
export function Adder({ inputRef, joining, onPick }: AdderProps) {
  const queryClient = useQueryClient();
  const [text, setText] = useState("");
  const typed = text.trim();
  const q = useDebouncedValue(typed, searchDebounceMs);

  // The movement library's query, so a movement added or renamed there shows up here too.
  const found = useQuery({
    queryKey: queryKeys.movements("list", q),
    queryFn: () => listMovements(q),
    enabled: q !== "",
  });

  // The chips stay until the next search is ready, or the row would collapse and reappear at each one. Not
  // keepPreviousData: that also brings back the search before a pick when the next one starts.
  const [shown, setShown] = useState<Movement[] | null>(null);
  const latest = !typed ? null : q && found.data ? found.data : shown;
  if (latest !== shown) setShown(latest);

  const create = useMutation({
    mutationFn: (name: string) => createMovement({ name, category: null, description: null }),
    onSuccess: (m) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.movements() });
      pick(m);
    },
  });

  const pick = (m: { id: string; name: string }) => {
    onPick(m);
    setText("");
    // Nothing in the new line is typed, so the keyboard would only cover it.
    inputRef.current?.blur();
  };

  const results = shown ?? [];
  const exact = results.some((m) => m.name === typed);
  // Keeps the keyboard up until the tap lands: closing it on press would move the chip from under the finger.
  const keepFocus = (e: { preventDefault: () => void }) => e.preventDefault();
  // iOS pans the page to reveal a tapped input, and the screen can only follow once the pan has ended, so the
  // search flies up and snaps back. Focusing it ourselves skips the pan; once focused, a tap just places the caret.
  const focusWithoutPan = (e: { preventDefault: () => void }) => {
    if (document.activeElement === inputRef.current) return;
    e.preventDefault();
    inputRef.current?.focus({ preventScroll: true });
  };

  return (
    <div className={s.adder}>
      {typed && (
        <div className={s.sug}>
          {results.slice(0, maxSuggestions).map((m) => (
            <button key={m.id} type="button" onPointerDown={keepFocus} onClick={() => pick(m)}>
              {m.name}
            </button>
          ))}
          {(shown || found.isError) && !exact && (
            <button
              type="button"
              className={s.create}
              onPointerDown={keepFocus}
              // Until the search for this exact text is back, it may yet find the name already in the library.
              disabled={create.isPending || q !== typed || found.isFetching}
              onClick={() => create.mutate(typed)}
            >
              + افزودن «{typed}»
            </button>
          )}
        </div>
      )}
      <label className={s.field} onTouchEnd={focusWithoutPan}>
        <SearchIcon />
        <input
          ref={inputRef}
          type="search"
          enterKeyHint="done"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={joining ? "حرکت بعدی سوپرست را جستجو کنید…" : "جستجو یا افزودن حرکت…"}
          aria-label="جستجوی حرکت"
        />
      </label>
      {create.isError && (
        <p className={s.lineError} role="alert">
          {create.error.message}
        </p>
      )}
    </div>
  );
}
