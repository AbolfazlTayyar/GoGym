import type { InputHTMLAttributes, ReactNode, TextareaHTMLAttributes } from "react";
import f from "./Field.module.css";

type FieldFrameProps = {
  label: string;
  optional?: boolean;
  error?: string;
};

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

type TextAreaFieldProps = TextareaHTMLAttributes<HTMLTextAreaElement> & FieldFrameProps;

// The boards' taller rounded field (.in.area) for a sentence or two, such as injuries or a description.
export function TextAreaField({ label, optional, error, ...textarea }: TextAreaFieldProps) {
  return (
    <label className={f.fl}>
      <span className={f.label}>
        {label}
        {optional && <i className={f.optional}> · اختیاری</i>}
      </span>
      <span className={error ? `${f.in} ${f.area} ${f.invalid}` : `${f.in} ${f.area}`}>
        <textarea className={f.input} aria-invalid={error ? true : undefined} {...textarea} />
      </span>
      {error && <span className={f.error}>{error}</span>}
    </label>
  );
}
