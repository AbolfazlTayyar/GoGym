import type { ReactNode } from "react";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import s from "./ConfirmSheet.module.css";

type ConfirmSheetProps = {
  open: boolean;
  onClose: () => void;
  title: ReactNode;
  text?: ReactNode;
  confirmLabel: string;
  onConfirm: () => void;
  pending?: boolean;
  error?: string | null;
};

// Asks before something that can't be undone; the pink button does it, the outline one backs out.
export function ConfirmSheet({
  open,
  onClose,
  title,
  text,
  confirmLabel,
  onConfirm,
  pending,
  error,
}: ConfirmSheetProps) {
  return (
    <BottomSheet open={open} onClose={onClose} title={title}>
      {text && <p className={s.text}>{text}</p>}
      <div className={s.acts}>
        <BigButton onClick={onConfirm} disabled={pending}>
          {confirmLabel}
        </BigButton>
        <BigButton outline onClick={onClose}>
          انصراف
        </BigButton>
      </div>
      {error && (
        <p className={s.error} role="alert">
          {error}
        </p>
      )}
    </BottomSheet>
  );
}
