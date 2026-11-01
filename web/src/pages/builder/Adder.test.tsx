// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { createRef } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Adder } from "./Adder";

const squat = { id: "m1", name: "اسکوات" };
const press = { id: "m2", name: "پرس سینه" };

// Makes the next search wait until the test answers it.
function deferredSearch() {
  let resolve: (movements: unknown[]) => void = () => {};
  const answer = new Promise<Response>((done) => {
    resolve = (movements) =>
      done(new Response(JSON.stringify({ success: true, data: movements, error: null })));
  });
  vi.mocked(fetch).mockImplementationOnce(() => answer);
  return { resolve };
}

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
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
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

  it("searches once for a burst of typing, not at each character", async () => {
    const input = renderAdder();

    fireEvent.change(input, { target: { value: "ا" } });
    fireEvent.change(input, { target: { value: "اس" } });
    fireEvent.change(input, { target: { value: "اسک" } });
    await screen.findByRole("button", { name: squat.name });

    expect(fetch).toHaveBeenCalledTimes(1);
    expect(vi.mocked(fetch).mock.calls[0][0]).toContain(encodeURIComponent("اسک"));
  });

  it("keeps the chips on screen while the next search loads", async () => {
    const input = renderAdder();
    fireEvent.change(input, { target: { value: "اسک" } });
    await screen.findByRole("button", { name: squat.name });

    const next = deferredSearch();
    fireEvent.change(input, { target: { value: "اسکو" } });
    await vi.waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    expect(screen.getByRole("button", { name: squat.name })).toBeTruthy();
    const add = screen.getByRole<HTMLButtonElement>("button", { name: "+ افزودن «اسکو»" });
    expect(add.disabled).toBe(true);

    next.resolve([press]);

    expect(await screen.findByRole("button", { name: press.name })).toBeTruthy();
    expect(screen.queryByRole("button", { name: squat.name })).toBeNull();
    expect(add.disabled).toBe(false);
  });

  it("starts the next search empty, without the chips from before a pick", async () => {
    const input = renderAdder();
    fireEvent.change(input, { target: { value: "اسک" } });
    fireEvent.click(await screen.findByRole("button", { name: squat.name }));

    const next = deferredSearch();
    // Long enough for the cleared text to settle, as it has by the time a coach starts the next line.
    await new Promise((resolve) => setTimeout(resolve, 450));
    fireEvent.change(input, { target: { value: "پرس" } });
    await vi.waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    expect(screen.queryByRole("button", { name: squat.name })).toBeNull();

    next.resolve([press]);
    expect(await screen.findByRole("button", { name: press.name })).toBeTruthy();
  });

  it("offers adding the name when the search fails", async () => {
    vi.mocked(fetch).mockImplementation(async () => new Response("nope", { status: 500 }));
    const input = renderAdder();

    fireEvent.change(input, { target: { value: "اسک" } });

    expect(await screen.findByRole("button", { name: "+ افزودن «اسک»" })).toBeTruthy();
  });
});
