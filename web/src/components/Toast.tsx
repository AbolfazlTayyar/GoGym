import { useEffect, useState } from "react";
import s from "./Toast.module.css";

const toastMs = 2500;
// Long enough to read the message and reach the button.
const actionToastMs = 5000;
// Matches the sink animation in Toast.module.css.
const leaveMs = 250;

type ToastAction = { label: string; onClick: () => void };

// useToast gives a screen one short message at a time; a new one replaces the one showing.
export function useToast() {
  const [message, setMessage] = useState<{
    text: string;
    action?: ToastAction;
    id: number;
    leaving?: boolean;
  } | null>(null);

  // A timer, not animationend, removes it after sinking, so it can never stay stuck on screen.
  useEffect(() => {
    if (!message) return;
    const timer = message.leaving
      ? setTimeout(() => setMessage(null), leaveMs)
      : setTimeout(
          () => setMessage({ ...message, leaving: true }),
          message.action ? actionToastMs : toastMs,
        );
    return () => clearTimeout(timer);
  }, [message]);

  // An action, such as undo, closes the toast when tapped.
  const show = (text: string, action?: ToastAction) => setMessage({ text, action, id: Date.now() });

  // The status region stays mounted, so screen readers announce each message put into it.
  const toast = (
    <div role="status" className={s.region}>
      {message && (
        <div key={message.id} className={message.leaving ? `${s.toast} ${s.leaving}` : s.toast}>
          {message.text}
          {message.action && (
            <button
              type="button"
              className={s.action}
              onClick={() => {
                setMessage({ ...message, leaving: true });
                message.action!.onClick();
              }}
            >
              {message.action.label}
            </button>
          )}
        </div>
      )}
    </div>
  );

  return { toast, show };
}
