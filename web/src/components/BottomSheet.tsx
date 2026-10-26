import {
  useEffect,
  useRef,
  useState,
  type AnimationEvent,
  type PointerEvent,
  type ReactNode,
} from "react";
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
  // What the panel shows, kept after open turns false so it slides down with its content: callers
  // render their form only while open.
  const [shown, setShown] = useState<{ children: ReactNode } | null>(null);
  if (open && shown?.children !== children) setShown({ children });
  const closing = !open && shown !== null;

  useEffect(() => {
    const dialog = ref.current;
    if (open && dialog && !dialog.open) dialog.showModal();
  }, [open]);

  // The dialog closes once the panel has slid down, not when open turns false.
  const onAnimationEnd = (e: AnimationEvent<HTMLDivElement>) => {
    if (!closing || e.target !== e.currentTarget) return;
    ref.current?.close();
    setShown(null);
    setDrag(0);
  };

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
    // Closing keeps the drag, so the panel slides down from where the finger left it.
    if (drag > swipeCloseDistance) onClose();
    else setDrag(0);
  };

  return (
    <dialog
      ref={ref}
      className={closing ? `${s.dialog} ${s.closing}` : s.dialog}
      onCancel={(e) => {
        // Escape: close through onClose so the history entry goes too.
        e.preventDefault();
        onClose();
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      {shown && (
        <div
          className={s.panel}
          style={drag ? { transform: `translateY(${drag}px)` } : undefined}
          onAnimationEnd={onAnimationEnd}
        >
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
          {shown.children}
        </div>
      )}
    </dialog>
  );
}
