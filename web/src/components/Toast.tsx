import { useEffect, useState } from "react";
import s from "./Toast.module.css";

const toastMs = 2500;

// useToast gives a screen one short message at a time; a new one replaces the one showing.
export function useToast() {
  const [message, setMessage] = useState<{ text: string; id: number } | null>(null);

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), toastMs);
    return () => clearTimeout(timer);
  }, [message]);

  const show = (text: string) => setMessage({ text, id: Date.now() });

  // The status region stays mounted, so screen readers announce each message put into it.
  const toast = (
    <div role="status" className={s.region}>
      {message && (
        <div key={message.id} className={s.toast}>
          {message.text}
        </div>
      )}
    </div>
  );

  return { toast, show };
}
