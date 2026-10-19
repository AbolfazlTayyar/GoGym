import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, type RefObject } from "react";
import { queryKeys } from "../../api/keys";
import { createMovement, listMovements } from "../../api/movements";
import { SearchIcon } from "../../components/icons";
import { useDebouncedValue } from "../../lib/useDebouncedValue";
import s from "./Builder.module.css";

const searchDebounceMs = 200;
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
  const q = useDebouncedValue(text.trim(), searchDebounceMs);

  // The movement library's query, so a movement added or renamed there shows up here too.
  const found = useQuery({
    queryKey: queryKeys.movements("list", q),
    queryFn: () => listMovements(q),
    enabled: q !== "",
  });
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
  };

  const typed = text.trim();
  const results = q && typed ? (found.data ?? []).slice(0, maxSuggestions) : [];
  const exact = results.some((m) => m.name === typed);
  // Keeps the input focused, and the keyboard up, when a chip is tapped.
  const keepFocus = (e: { preventDefault: () => void }) => e.preventDefault();

  return (
    <div className={s.adder}>
      {typed && (
        <div className={s.sug}>
          {results.map((m) => (
            <button key={m.id} type="button" onPointerDown={keepFocus} onClick={() => pick(m)}>
              {m.name}
            </button>
          ))}
          {!exact && q === typed && !found.isFetching && (
            <button
              type="button"
              className={s.create}
              onPointerDown={keepFocus}
              disabled={create.isPending}
              onClick={() => create.mutate(typed)}
            >
              + افزودن «{typed}»
            </button>
          )}
        </div>
      )}
      <label className={s.field}>
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
