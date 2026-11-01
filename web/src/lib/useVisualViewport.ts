import { useEffect, useState, type CSSProperties } from "react";

// iOS Safari doesn't shrink the layout for the on-screen keyboard, so a bar pinned to the bottom ends up
// under it. Sizing the screen to the visual viewport keeps it above. Android Chrome resizes the layout
// itself (interactive-widget=resizes-content in index.html), so it must stay on 100dvh: a pixel height
// lags the resize by a frame, the page is briefly too tall, and Chrome scrolls it up to reach the input.
// Moves with the iOS keyboard, ease-out so the search stays ahead of it while opening. Only here, not in the
// Screen's CSS: Android's 100dvh changes with the keyboard and must never lag behind it.
const followKeyboard = "height 250ms ease-out";
const closing: CSSProperties = { transition: followKeyboard };

export function useVisualViewportStyle(): CSSProperties | undefined {
  const [style, setStyle] = useState<CSSProperties>();

  useEffect(() => {
    const vv = window.visualViewport;
    if (!vv) return;
    const update = () => {
      // Closing, iOS keeps reporting the keyboard's height until it has fully gone, which would leave the screen
      // short over an empty space and then drop it. Focus leaves the input as the keyboard starts to close, so
      // the screen grows back from there, keeping the transition until the viewport has caught up.
      const typing = document.activeElement?.matches("input, textarea");
      // clientHeight, not innerHeight: iOS can shrink innerHeight with the keyboard, which reads as a resized
      // layout. vv.height is fractional and clientHeight a whole number, so a gap under a pixel is only rounding.
      const layoutTallerThanVisible = document.documentElement.clientHeight - vv.height > 1;
      setStyle((prev) => {
        if (!layoutTallerThanVisible) return undefined;
        if (typing) {
          return {
            height: `${vv.height}px`,
            transform: `translateY(${vv.offsetTop}px)`,
            transition: followKeyboard,
          };
        }
        return prev && closing;
      });
    };
    update();
    vv.addEventListener("resize", update);
    vv.addEventListener("scroll", update);
    document.addEventListener("focusout", update);
    return () => {
      vv.removeEventListener("resize", update);
      vv.removeEventListener("scroll", update);
      document.removeEventListener("focusout", update);
    };
  }, []);

  return style;
}
