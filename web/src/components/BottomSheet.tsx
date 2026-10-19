import { useEffect, useRef, useState, type PointerEvent, type ReactNode } from "react";
import { RoundButton } from "./Header";
import { XIcon } from "./icons";
import s from "./BottomSheet.module.css";

const swipeCloseDistance = 80;

type BottomSheetProps = {
  open: boolean;
  onClose: () => void;
  title: ReactNode;
  children: ReactNode;
};

// A native modal <dialog>, so focus and the inert page come for free. Open it through useSheet, which
// gives it a history entry: the phone's back button then closes the sheet instead of leaving the page.
export function BottomSheet({ open, onClose, title, children }: BottomSheetProps) {
  const ref = useRef<HTMLDialogElement>(null);
  const dragStart = useRef<number | null>(null);
  const [drag, setDrag] = useState(0);

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);

  const onPointerDown = (e: PointerEvent<HTMLDivElement>) => {
    if ((e.target as Element).closest("button")) return;
    dragStart.current = e.clientY;
    e.currentTarget.setPointerCapture(e.pointerId);
  };
  const onPointerMove = (e: PointerEvent<HTMLDivElement>) => {
    if (dragStart.current !== null) setDrag(Math.max(0, e.clientY - dragStart.current));
  };
  const onPointerUp = () => {
    if (dragStart.current === null) return;
    dragStart.current = null;
    if (drag > swipeCloseDistance) onClose();
    setDrag(0);
  };

  return (
    <dialog
      ref={ref}
      className={s.dialog}
      onCancel={(e) => {
        // Escape: close through onClose so the history entry goes too.
        e.preventDefault();
        onClose();
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      {open && (
        <div className={s.panel} style={drag ? { transform: `translateY(${drag}px)` } : undefined}>
          <div
            className={s.handle}
            onPointerDown={onPointerDown}
            onPointerMove={onPointerMove}
            onPointerUp={onPointerUp}
            onPointerCancel={onPointerUp}
          >
            <div className={s.grab} />
            <div className={s.bsh}>
              <b>{title}</b>
              <RoundButton label="بستن" onClick={onClose}>
                <XIcon />
              </RoundButton>
            </div>
          </div>
          {children}
        </div>
      )}
    </dialog>
  );
}
