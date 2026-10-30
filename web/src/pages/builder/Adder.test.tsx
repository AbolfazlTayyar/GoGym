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

describe("builder adder", () => {
  it("closes the keyboard and clears the search after a pick", async () => {
    const onPick = vi.fn();
    render(
      <QueryClientProvider client={new QueryClient()}>
        <Adder inputRef={createRef<HTMLInputElement>()} joining={false} onPick={onPick} />
      </QueryClientProvider>,
    );
    const input = screen.getByRole<HTMLInputElement>("searchbox");
    input.focus();
    fireEvent.change(input, { target: { value: squat.name } });

    fireEvent.click(await screen.findByRole("button", { name: squat.name }));

    expect(onPick).toHaveBeenCalledWith(squat);
    expect(input.value).toBe("");
    expect(document.activeElement).not.toBe(input);
  });
});
