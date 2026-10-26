// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { LoginPage } from "./LoginPage";

const fetchMock = vi.fn();

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  cleanup();
  fetchMock.mockReset();
  vi.unstubAllGlobals();
});

describe("login", () => {
  it("names each empty field without calling the API", async () => {
    // GET /coaches/me: not signed in, so the form shows.
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          success: false,
          data: null,
          error: { code: "unauthorized", message: "unauthorized" },
        }),
        { status: 401 },
      ),
    );
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/login"]}>
          <LoginPage />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    fireEvent.click(await screen.findByRole("button", { name: "ورود" }));

    expect(screen.getByText("شماره موبایل را وارد کنید.")).toBeTruthy();
    expect(screen.getByText("رمز عبور را وارد کنید.")).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
