// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Movement } from "../api/types";
import { CategoriesSheet } from "./CategoriesSheet";

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

function movement(name: string, category: string, coachId: string | null = "c1"): Movement {
  return {
    id: name,
    coach_id: coachId,
    name,
    category,
    description: null,
    media_url: null,
    created_at: "",
    updated_at: "",
  };
}

const library = [
  movement("Goblet squat", "پا"),
  movement("Lunge", "پا"),
  movement("Plank", "مرکزی"),
  movement("Back squat", "قدرتی", null),
];

function renderSheet() {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter>
        <CategoriesSheet open onClose={() => {}} movements={library} />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("categories sheet", () => {
  it("lists the categories, built-in ones locked", () => {
    renderSheet();

    expect(screen.getByRole("button", { name: /پا/ })).toBeTruthy();
    expect(screen.queryByRole("button", { name: /قدرتی/ })).toBeNull();
    expect(screen.getByText("قدرتی")).toBeTruthy();
  });

  it("opens at the list again after it was closed on a category", () => {
    const client = new QueryClient();
    const sheet = (open: boolean) => (
      <QueryClientProvider client={client}>
        <MemoryRouter>
          <CategoriesSheet open={open} onClose={() => {}} movements={library} />
        </MemoryRouter>
      </QueryClientProvider>
    );
    const { rerender } = render(sheet(true));
    fireEvent.click(screen.getByRole("button", { name: /پا/ }));
    expect(screen.getByLabelText("نام دسته")).toBeTruthy();

    // Reopened before the sheet finished sliding down, which jsdom never animates anyway.
    rerender(sheet(false));
    rerender(sheet(true));

    expect(screen.queryByLabelText("نام دسته")).toBeNull();
    expect(screen.getByRole("button", { name: /مرکزی/ })).toBeTruthy();
  });

  it("names the empty name without calling the API", () => {
    renderSheet();
    fireEvent.click(screen.getByRole("button", { name: /پا/ }));

    fireEvent.change(screen.getByLabelText("نام دسته"), { target: { value: " ‌ " } });
    fireEvent.click(screen.getByRole("button", { name: "ذخیره" }));

    expect(screen.getByText("نام دسته را وارد کنید.")).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("warns before a rename merges into a category in use, matching an Arabic spelling", () => {
    renderSheet();
    fireEvent.click(screen.getByRole("button", { name: /مرکزی/ }));

    fireEvent.change(screen.getByLabelText("نام دسته"), { target: { value: "قدرتى" } });
    fireEvent.click(screen.getByRole("button", { name: "ذخیره" }));

    expect(screen.getByText("دسته «مرکزی» با «قدرتی» یکی می‌شود.")).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("swipes a category to delete it, asking first", () => {
    renderSheet();

    // The actions behind each of the coach's rows, in list order: پا, then مرکزی.
    fireEvent.click(screen.getAllByRole("button", { name: "حذف", hidden: true })[1]);

    expect(screen.getByText("حذف دسته «مرکزی»؟")).toBeTruthy();
    expect(screen.getByText("۱ حرکت بدون دسته می‌شوند.")).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("asks before deleting, naming the movements it leaves without a category", () => {
    renderSheet();
    fireEvent.click(screen.getByRole("button", { name: /پا/ }));

    fireEvent.click(screen.getByRole("button", { name: "حذف" }));

    expect(screen.getByText("حذف دسته «پا»؟")).toBeTruthy();
    expect(screen.getByText("۲ حرکت بدون دسته می‌شوند.")).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
