import { useEffect, useLayoutEffect, type RefObject } from "react";
import { useLocation } from "react-router";

// Scroll positions by history entry. The list scrolls inside its own box, which the router's
// window-based scroll restoration doesn't see.
const positions = new Map<string, number>();

// Back to an entry puts its list where the coach left it, once `ready` says the rows are rendered.
export function useScrollMemory(ref: RefObject<HTMLElement | null>, ready: boolean) {
  const { key } = useLocation();

  useLayoutEffect(() => {
    if (ready && ref.current) ref.current.scrollTop = positions.get(key) ?? 0;
  }, [ref, key, ready]);

  useEffect(() => {
    const el = ref.current;
    if (!ready || !el) return;
    const save = () => positions.set(key, el.scrollTop);
    el.addEventListener("scroll", save, { passive: true });
    return () => el.removeEventListener("scroll", save);
  }, [ref, key, ready]);
}
