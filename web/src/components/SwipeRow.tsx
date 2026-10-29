import { useRef, useState, type MouseEvent, type PointerEvent, type ReactNode } from "react";
import { EditIcon, TrashIcon } from "./icons";
import s from "./SwipeRow.module.css";

// Below this a touch is still a tap, and the first axis past it decides between swipe and scroll.
const dragThreshold = 8;

type SwipeRowProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEdit: () => void;
  onDelete: () => void;
  className?: string;
  children: ReactNode;
};

// A row that slides left under the finger, as in iOS Messages, to show edit and delete behind its right
// edge. The app is RTL, so that edge is inline-start.
export function SwipeRow({
  open,
  onOpenChange,
  onEdit,
  onDelete,
  className,
  children,
}: SwipeRowProps) {
  const actionsRef = useRef<HTMLDivElement>(null);
  const gesture = useRef<{
    x: number;
    y: number;
    width: number;
    from: number;
    horizontal: boolean | null;
  } | null>(null);
  // Set by a swipe, so the click that ends it doesn't follow the row's link.
  const swiped = useRef(false);
  const [width, setWidth] = useState(0);
  const [drag, setDrag] = useState<number | null>(null);
  const offset = drag ?? (open ? width : 0);

  const onPointerDown = (e: PointerEvent<HTMLDivElement>) => {
    if (e.button !== 0 || actionsRef.current?.contains(e.target as Node)) return;
    const w = actionsRef.current?.offsetWidth ?? 0;
    setWidth(w);
    gesture.current = {
      x: e.clientX,
      y: e.clientY,
      width: w,
      from: open ? w : 0,
      horizontal: null,
    };
    swiped.current = false;
  };

  const onPointerMove = (e: PointerEvent<HTMLDivElement>) => {
    const g = gesture.current;
    if (!g) return;
    const dx = g.x - e.clientX;
    if (g.horizontal === null) {
      const dy = e.clientY - g.y;
      if (Math.abs(dx) < dragThreshold && Math.abs(dy) < dragThreshold) return;
      g.horizontal = Math.abs(dx) > Math.abs(dy);
      if (!g.horizontal) return;
      e.currentTarget.setPointerCapture(e.pointerId);
      swiped.current = true;
    }
    if (g.horizontal) setDrag(Math.min(g.width, Math.max(0, g.from + dx)));
  };

  const onPointerUp = () => {
    const g = gesture.current;
    gesture.current = null;
    if (!g || drag === null) return;
    onOpenChange(drag > g.width / 2);
    setDrag(null);
  };

  // While open, a tap on the row closes it instead of opening it.
  const onClickCapture = (e: MouseEvent<HTMLDivElement>) => {
    if (actionsRef.current?.contains(e.target as Node)) return;
    if (!swiped.current && !open) return;
    e.preventDefault();
    e.stopPropagation();
    if (!swiped.current) onOpenChange(false);
    swiped.current = false;
  };

  return (
    <div
      className={className ? `${s.row} ${className}` : s.row}
      onPointerDown={onPointerDown}
      onPointerMove={onPointerMove}
      onPointerUp={onPointerUp}
      onPointerCancel={onPointerUp}
      onClickCapture={onClickCapture}
      onDragStart={(e) => e.preventDefault()}
    >
      {/* Hidden while closed, so they can't peek past the row's rounded corners or take focus. */}
      <div ref={actionsRef} className={offset > 0 ? s.actions : `${s.actions} ${s.hidden}`}>
        <button type="button" className={`${s.act} ${s.del}`} onClick={onDelete}>
          <TrashIcon />
          حذف
        </button>
        <button type="button" className={`${s.act} ${s.edit}`} onClick={onEdit}>
          <EditIcon />
          ویرایش
        </button>
      </div>
      <div
        className={drag === null ? s.content : `${s.content} ${s.dragging}`}
        style={offset ? { transform: `translateX(${-offset}px)` } : undefined}
      >
        {children}
      </div>
    </div>
  );
}
