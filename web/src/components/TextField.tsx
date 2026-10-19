import type { InputHTMLAttributes, ReactNode } from "react";
import f from "./Field.module.css";

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  optional?: boolean;
  error?: string;
  // Shown inside the pill after the input: a unit, or a button such as the password eye.
  trailing?: ReactNode;
  ltr?: boolean;
};

export function TextField({ label, optional, error, trailing, ltr, ...input }: TextFieldProps) {
  return (
    <label className={f.fl}>
      <span className={f.label}>
        {label}
        {optional && <i className={f.optional}> · اختیاری</i>}
      </span>
      <span className={error ? `${f.in} ${f.invalid}` : f.in}>
        <input
          className={ltr ? `${f.input} ${f.ltr}` : f.input}
          aria-invalid={error ? true : undefined}
          {...input}
        />
        {trailing}
      </span>
      {error && <span className={f.error}>{error}</span>}
    </label>
  );
}
