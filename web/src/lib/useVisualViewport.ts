import { useEffect, useState, type CSSProperties } from "react";

// iOS Safari doesn't shrink the layout for the on-screen keyboard, so a bar pinned to the bottom ends up
// under it. Sizing the screen to the visual viewport keeps it above. Android Chrome resizes the layout
// itself (interactive-widget=resizes-content in index.html), so it must stay on 100dvh: a pixel height
// lags the resize by a frame, the page is briefly too tall, and Chrome scrolls it up to reach the input.
export function useVisualViewportStyle(): CSSProperties | undefined {
  const [style, setStyle] = useState<CSSProperties>();

  useEffect(() => {
    const vv = window.visualViewport;
    if (!vv) return;
    const update = () => {
      // clientHeight, not innerHeight: iOS can shrink innerHeight with the keyboard, which reads as a resized
      // layout. vv.height is fractional and clientHeight a whole number, so a gap under a pixel is only rounding.
      const layoutTallerThanVisible = document.documentElement.clientHeight - vv.height > 1;
      setStyle(
        layoutTallerThanVisible
          ? { height: `${vv.height}px`, transform: `translateY(${vv.offsetTop}px)` }
          : undefined,
      );
    };
    update();
    vv.addEventListener("resize", update);
    vv.addEventListener("scroll", update);
    return () => {
      vv.removeEventListener("resize", update);
      vv.removeEventListener("scroll", update);
    };
  }, []);

  return style;
}
