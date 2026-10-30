// @vitest-environment jsdom
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useVisualViewportStyle } from "./useVisualViewport";

function stubViewport(innerHeight: number, height: number, offsetTop = 0) {
  const viewport = Object.assign(new EventTarget(), { height, offsetTop });
  vi.stubGlobal("innerHeight", innerHeight);
  vi.stubGlobal("visualViewport", viewport);
  return viewport;
}

describe("useVisualViewportStyle", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("follows the visual viewport while the layout is taller than it (iOS keyboard)", () => {
    stubViewport(800, 450, 120);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toEqual({ height: "450px", transform: "translateY(120px)" });
  });

  it("leaves the height to CSS when the layout resized with the keyboard (Android)", () => {
    stubViewport(450, 450);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toBeUndefined();
  });

  it("treats a gap under a pixel as rounding", () => {
    stubViewport(732, 731.43);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toBeUndefined();
  });

  it("drops the style when the iOS keyboard closes", () => {
    const viewport = stubViewport(800, 450);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toBeDefined();

    viewport.height = 800;
    act(() => {
      viewport.dispatchEvent(new Event("resize"));
    });
    expect(result.current).toBeUndefined();
  });
});
