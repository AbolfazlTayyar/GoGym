// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MovementSheet } from "./MovementSheet";

const fetchMock = vi.fn();

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
  // jsdom has no <dialog> methods; opening only needs the attribute.
  HTMLDialogElement.prototype.showModal ??= function (this: HTMLDialogElement) {
    this.setAttribute("open", "");
  };
  HTMLDialogElement.prototype.close ??= vi.fn();
});

afterEach(() => {
  cleanup();
  fetchMock.mockReset();
  vi.unstubAllGlobals();
});

describe("movement sheet", () => {
  it("names the empty name without calling the API", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter>
          <MovementSheet open onClose={() => {}} movement={null} categories={[]} />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "ذخیره" }));

    expect(screen.getByText("نام حرکت را وارد کنید.")).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
