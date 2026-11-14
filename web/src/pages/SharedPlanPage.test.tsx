// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SharedPlanPage } from "./SharedPlanPage";

const fetchMock = vi.fn();

const coach = { first_name: "علی", last_name: "رضایی", phone: "09121234567" };

function answer(status: number, body: object) {
  fetchMock.mockResolvedValueOnce(new Response(JSON.stringify(body), { status }));
}

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/p/q3J9xK2mT0aLwV7nB4cD8e"]}>
        <Routes>
          <Route path="/p/:token" element={<SharedPlanPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  cleanup();
  fetchMock.mockReset();
  vi.unstubAllGlobals();
});

describe("shared plan page", () => {
  it("shows the coach, the athlete and every day of an active link", async () => {
    const line = (name: string) => ({
      id: name,
      order_index: 0,
      sets: 3,
      rest_seconds: null,
      notes: null,
      movements: [
        {
          id: name,
          movement_id: name,
          name,
          category: null,
          reps: 10,
          duration_seconds: null,
          load_kg: null,
          order_in_block: 0,
        },
      ],
    });
    answer(200, {
      success: true,
      error: null,
      data: {
        status: "active",
        expires_at: "2026-12-14T10:00:00Z",
        coach,
        athlete: { first_name: "سارا", last_name: "محمدی" },
        plan: {
          title: "برنامه آبان",
          note: null,
          start_date: "2026-11-15",
          days: [
            { id: "a", label: "A", order_index: 0, blocks: [line("اسکات")] },
            { id: "b", label: "B", order_index: 1, blocks: [line("ددلیفت")] },
          ],
        },
      },
    });
    renderPage();

    expect(await screen.findByText("علی رضایی")).toBeTruthy();
    expect(screen.getByRole("link", { name: /۰۹۱۲/ }).getAttribute("href")).toBe("tel:09121234567");
    expect(screen.getByRole("heading", { name: "سارا محمدی" })).toBeTruthy();
    expect(screen.getByText("اسکات")).toBeTruthy();
    expect(screen.getByText("ددلیفت")).toBeTruthy();
    expect(screen.getByRole("button", { name: "دریافت PDF" })).toBeTruthy();
    expect(screen.getByText("ساخته‌شده با gogymapp")).toBeTruthy();
    expect(document.querySelector('meta[name="robots"]')?.getAttribute("content")).toBe("noindex");
  });

  it("still names the coach on an expired link, with no plan", async () => {
    answer(200, {
      success: true,
      error: null,
      data: {
        status: "expired",
        expires_at: "2026-10-14T10:00:00Z",
        coach,
        athlete: null,
        plan: null,
      },
    });
    renderPage();

    expect(await screen.findByText("این لینک منقضی شده است")).toBeTruthy();
    expect(screen.getByText("علی رضایی")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "دریافت PDF" })).toBeNull();
  });

  it("says a replaced or unknown link is no longer valid, with no coach", async () => {
    answer(404, {
      success: false,
      data: null,
      error: { code: "not_found", message: "not found" },
    });
    renderPage();

    expect(await screen.findByText("این لینک دیگر معتبر نیست")).toBeTruthy();
    expect(screen.queryByText("علی رضایی")).toBeNull();
    expect(fetchMock).toHaveBeenCalledOnce();
  });
});
