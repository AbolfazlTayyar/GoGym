// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { createRef } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Adder } from "./Adder";

const squat = { id: "m1", name: "اسکوات" };

beforeEach(() => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response(JSON.stringify({ success: true, data: [squat], error: null }))),
  );
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function renderAdder(onPick = vi.fn()) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <Adder inputRef={createRef<HTMLInputElement>()} joining={false} onPick={onPick} />
    </QueryClientProvider>,
  );
  return screen.getByRole<HTMLInputElement>("searchbox");
}

describe("builder adder", () => {
  it("focuses a tapped search itself, without letting the page pan to it", () => {
    const input = renderAdder();
    const focus = vi.spyOn(input, "focus");

    const notPrevented = fireEvent.touchEnd(input);

    expect(notPrevented).toBe(false);
    expect(focus).toHaveBeenCalledWith({ preventScroll: true });
    expect(document.activeElement).toBe(input);
  });

  it("leaves a tap on the focused search alone, so it can place the caret", () => {
    const input = renderAdder();
    input.focus();

    expect(fireEvent.touchEnd(input)).toBe(true);
  });

  it("closes the keyboard and clears the search after a pick", async () => {
    const onPick = vi.fn();
    const input = renderAdder(onPick);
    input.focus();
    fireEvent.change(input, { target: { value: squat.name } });

    fireEvent.click(await screen.findByRole("button", { name: squat.name }));

    expect(onPick).toHaveBeenCalledWith(squat);
    expect(input.value).toBe("");
    expect(document.activeElement).not.toBe(input);
  });
});
