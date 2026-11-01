// @vitest-environment jsdom
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useVisualViewportStyle } from "./useVisualViewport";

function stubViewport(layoutHeight: number, height: number, offsetTop = 0) {
  const viewport = Object.assign(new EventTarget(), { height, offsetTop });
  vi.spyOn(document.documentElement, "clientHeight", "get").mockReturnValue(layoutHeight);
  vi.stubGlobal("visualViewport", viewport);
  return viewport;
}

describe("useVisualViewportStyle", () => {
  let input: HTMLInputElement;

  beforeEach(() => {
    input = document.body.appendChild(document.createElement("input"));
    input.focus();
  });

  afterEach(() => {
    cleanup();
    input.remove();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("follows the visual viewport while the layout is taller than it (iOS keyboard)", () => {
    stubViewport(800, 450, 120);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toEqual({
      height: "450px",
      transform: "translateY(120px)",
      transition: expect.stringContaining("height"),
    });
  });

  it("follows it on an iOS whose innerHeight shrinks with the keyboard", () => {
    stubViewport(800, 450);
    vi.stubGlobal("innerHeight", 450);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toMatchObject({ height: "450px", transform: "translateY(0px)" });
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

  it("animates back to full height as focus leaves, while iOS still reports the keyboard's height", () => {
    const viewport = stubViewport(800, 450);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current?.height).toBe("450px");

    const closing = { transition: expect.stringContaining("height") };
    act(() => input.blur());
    expect(result.current).toEqual(closing);

    act(() => {
      viewport.dispatchEvent(new Event("scroll"));
    });
    expect(result.current).toEqual(closing);

    viewport.height = 800;
    act(() => {
      viewport.dispatchEvent(new Event("resize"));
    });
    expect(result.current).toBeUndefined();
  });

  it("ignores a shrunken viewport with nothing to type in (pinch zoom)", () => {
    input.blur();
    stubViewport(800, 450);
    const { result } = renderHook(() => useVisualViewportStyle());
    expect(result.current).toBeUndefined();
  });
});
