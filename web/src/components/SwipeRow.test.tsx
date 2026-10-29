// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SwipeRow } from "./SwipeRow";

const actionsWidth = 146;

function Harness({ onRowClick }: { onRowClick: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <SwipeRow open={open} onOpenChange={setOpen} onEdit={() => {}} onDelete={() => {}}>
      <button type="button" onClick={onRowClick}>
        برنامه مهر
      </button>
    </SwipeRow>
  );
}

function swipe(el: HTMLElement, fromX: number, toX: number) {
  fireEvent.pointerDown(el, { button: 0, clientX: fromX, clientY: 100, pointerId: 1 });
  fireEvent.pointerMove(el, { clientX: (fromX + toX) / 2, clientY: 100, pointerId: 1 });
  fireEvent.pointerMove(el, { clientX: toX, clientY: 100, pointerId: 1 });
  fireEvent.pointerUp(el, { clientX: toX, clientY: 100, pointerId: 1 });
  fireEvent.click(el);
}

// jsdom has no layout or pointer capture.
beforeEach(() => {
  vi.spyOn(HTMLElement.prototype, "offsetWidth", "get").mockReturnValue(actionsWidth);
  Element.prototype.setPointerCapture = () => {};
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

const content = () => screen.getByText("برنامه مهر").parentElement!;

describe("SwipeRow", () => {
  it("opens on a swipe to the left past half the actions, without following the row", () => {
    const onRowClick = vi.fn();
    render(<Harness onRowClick={onRowClick} />);
    const row = screen.getByText("برنامه مهر");

    swipe(row, 300, 200);

    expect(content().style.transform).toBe(`translateX(-${actionsWidth}px)`);
    expect(onRowClick).not.toHaveBeenCalled();
  });

  it("springs back on a short swipe", () => {
    render(<Harness onRowClick={() => {}} />);
    swipe(screen.getByText("برنامه مهر"), 300, 250);

    expect(content().style.transform).toBe("");
  });

  it("closes on a tap while open, and a tap while closed follows the row", () => {
    const onRowClick = vi.fn();
    render(<Harness onRowClick={onRowClick} />);
    const row = screen.getByText("برنامه مهر");
    swipe(row, 300, 100);

    fireEvent.pointerDown(row, { button: 0, clientX: 150, clientY: 100, pointerId: 1 });
    fireEvent.pointerUp(row, { clientX: 150, clientY: 100, pointerId: 1 });
    fireEvent.click(row);
    expect(content().style.transform).toBe("");
    expect(onRowClick).not.toHaveBeenCalled();

    fireEvent.pointerDown(row, { button: 0, clientX: 150, clientY: 100, pointerId: 1 });
    fireEvent.pointerUp(row, { clientX: 150, clientY: 100, pointerId: 1 });
    fireEvent.click(row);
    expect(onRowClick).toHaveBeenCalledOnce();
  });

  it("leaves a vertical drag to the page's scroll", () => {
    render(<Harness onRowClick={() => {}} />);
    const row = screen.getByText("برنامه مهر");

    fireEvent.pointerDown(row, { button: 0, clientX: 300, clientY: 100, pointerId: 1 });
    fireEvent.pointerMove(row, { clientX: 290, clientY: 160, pointerId: 1 });
    fireEvent.pointerMove(row, { clientX: 200, clientY: 170, pointerId: 1 });

    expect(content().style.transform).toBe("");
  });
});
