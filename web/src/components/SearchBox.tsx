import { SearchIcon, XIcon } from "./icons";
import s from "./SearchBox.module.css";

type SearchBoxProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
};

export function SearchBox({ value, onChange, placeholder }: SearchBoxProps) {
  return (
    <div className={value ? `${s.search} ${s.on}` : s.search}>
      <SearchIcon />
      <input
        type="search"
        enterKeyHint="search"
        className={s.input}
        value={value}
        placeholder={placeholder}
        aria-label={placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
      {value && (
        <button type="button" className={s.clr} aria-label="پاک کردن" onClick={() => onChange("")}>
          <XIcon />
        </button>
      )}
    </div>
  );
}
