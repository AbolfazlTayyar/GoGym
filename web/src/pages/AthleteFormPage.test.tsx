// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { NewAthletePage } from "./AthleteFormPage";

const fetchMock = vi.fn();

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
  // jsdom has no <dialog>; the form's sheets stay closed in these tests.
  HTMLDialogElement.prototype.showModal ??= vi.fn();
  HTMLDialogElement.prototype.close ??= vi.fn();
});

afterEach(() => {
  cleanup();
  fetchMock.mockReset();
  vi.unstubAllGlobals();
});

function renderForm() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/athletes/new?type=private"]}>
        <Routes>
          <Route path="/athletes/new" element={<NewAthletePage />} />
          <Route path="/athletes/:id" element={<p>profile</p>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

// Labels hold more than their name (the optional marker, a unit), so they match from the start.
const type = (label: string, value: string) =>
  fireEvent.change(screen.getByLabelText(new RegExp(`^${label}( ·|$|سانتی)`)), {
    target: { value },
  });

describe("athlete form", () => {
  it("names each empty required field without calling the API", () => {
    renderForm();

    fireEvent.click(screen.getByRole("button", { name: "ذخیره" }));

    expect(screen.getByText("نام را وارد کنید.")).toBeTruthy();
    expect(screen.getByText("نام خانوادگی را وارد کنید.")).toBeTruthy();
    expect(screen.getByText("شماره موبایل را وارد کنید.")).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("shows the API's phone error under the mobile field", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          success: false,
          data: null,
          error: {
            code: "validation_failed",
            message: "invalid request",
            fields: { phone: "must be an Iranian mobile number, e.g. 09372144430" },
          },
        }),
        { status: 400 },
      ),
    );
    renderForm();
    type("نام", "سارا");
    type("نام خانوادگی", "کریمی");
    type("موبایل", "123");

    fireEvent.click(screen.getByRole("button", { name: "ذخیره" }));

    const message = await screen.findByText("شماره موبایل درست نیست؛ مثلاً ۰۹۱۲۳۴۵۶۷۸۹.");
    expect(message.closest("label")?.textContent).toContain("موبایل");
  });

  it("sends Persian digits in mobile and height as Latin", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ success: true, data: { id: "a1" }, error: null }), {
        status: 201,
      }),
    );
    renderForm();
    type("نام", "سارا");
    type("نام خانوادگی", "کریمی");
    type("موبایل", "۰۹۱۲ ۳۴۵ ۶۷۸۹");
    type("قد", "۱۶۸٫۵");

    fireEvent.click(screen.getByRole("button", { name: "ذخیره" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalled());
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(body).toMatchObject({
      first_name: "سارا",
      phone: "09123456789",
      height: 168.5,
      athlete_type: "private",
      experience_level: null,
      goal: null,
    });
    expect(await screen.findByText("profile")).toBeTruthy();
  });
});
