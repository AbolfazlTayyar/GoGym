// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { Athlete, AttendanceCount, PlanSummary } from "../api/types";
import { todayISO } from "../lib/jalali";
import { AthleteRow } from "./AthleteRow";

const athlete: Athlete = {
  id: "a1",
  first_name: "سارا",
  last_name: "کریمی",
  phone: "09123456789",
  experience_level: null,
  injuries: null,
  goal: null,
  height: null,
  athlete_type: "private",
  created_at: "2026-09-01T00:00:00Z",
  updated_at: "2026-09-01T00:00:00Z",
};

const plan = (labels: string[]): PlanSummary => ({
  id: "p1",
  athlete_id: "a1",
  title: "برنامه مهر",
  note: null,
  start_date: "2026-09-23",
  is_current: true,
  day_labels: labels,
  block_count: 40,
  created_at: "2026-09-01T00:00:00Z",
  updated_at: "2026-09-01T00:00:00Z",
});

function renderRow(p: PlanSummary | null) {
  return render(
    <MemoryRouter>
      <AthleteRow athlete={athlete} plan={p} columns={5} />
    </MemoryRouter>,
  );
}

afterEach(cleanup);

describe("AthleteRow", () => {
  it("shows the no-plan pill and a new-plan button when there is no plan", () => {
    renderRow(null);

    expect(screen.getByText("بدون برنامه")).toBeTruthy();
    expect(screen.getByRole("link", { name: /برنامه جدید/ }).getAttribute("href")).toBe(
      "/athletes/a1/plans/new",
    );
    expect(screen.getByRole("link", { name: /سارا کریمی/ }).getAttribute("href")).toBe(
      "/athletes/a1",
    );
  });

  it("leaves the fifth place empty for a four-day plan", () => {
    const { container } = renderRow(plan(["A", "B", "C", "D"]));

    const letters = screen.getAllByRole("link").filter((a) => /^[A-G]$/.test(a.textContent ?? ""));
    expect(letters.map((a) => a.textContent)).toEqual(["A", "B", "C", "D"]);
    expect(letters[1].getAttribute("href")).toBe("/plans/p1?day=B");
    expect(container.querySelectorAll('[data-testid="empty-day"]')).toHaveLength(1);
    const cells = container.querySelector('[data-testid="empty-day"]')!.parentElement!.children;
    expect(cells[4].getAttribute("data-testid")).toBe("empty-day");
    expect(screen.queryByText("بدون برنامه")).toBeNull();
  });

  describe("sessions", () => {
    const count = (visits: number, lastDate: string | null): AttendanceCount => ({
      plan_id: "p1",
      athlete_id: "a1",
      visits,
      sessions: 12,
      last_date: lastDate,
    });

    function renderSessions(c: AttendanceCount, onVisit = () => {}) {
      return render(
        <MemoryRouter>
          <AthleteRow
            athlete={athlete}
            plan={plan(["A", "B"])}
            columns={5}
            sessions={{ count: c, onVisit, pending: false }}
          />
        </MemoryRouter>,
      );
    }

    it("records today's visit with one tap", () => {
      const onVisit = vi.fn();
      renderSessions(count(7, "2026-01-01"), onVisit);

      fireEvent.click(screen.getByRole("button", { name: /ثبت حضور امروز/ }));

      expect(onVisit).toHaveBeenCalledOnce();
      expect(screen.getByText("+ ۷ از ۱۲")).toBeTruthy();
    });

    it("offers no tap once today is recorded", () => {
      renderSessions(count(7, todayISO()));

      expect(screen.queryByRole("button")).toBeNull();
      expect(screen.getByText("✓ ۷ از ۱۲")).toBeTruthy();
    });

    it("says the sessions are used after the 12th", () => {
      renderSessions(count(12, todayISO()));

      expect(screen.queryByRole("button")).toBeNull();
      expect(screen.getByText("تمام شد")).toBeTruthy();
    });
  });
});
