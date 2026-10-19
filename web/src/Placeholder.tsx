import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { api } from "./api/client";
import f from "./components/Field.module.css";
import { JalaliDateField } from "./components/JalaliDateField";
import { parseNumber } from "./lib/digits";
import { todayISO } from "./lib/jalali";

// A dev check page: the proxy reaches the API, API errors read in Persian, and Persian input parses.
// The sign-in screen replaces it.
export function Placeholder() {
  const health = useQuery({
    queryKey: ["healthz"],
    queryFn: async () => {
      const res = await fetch("/healthz");
      const body: { status: string } = await res.json();
      return body.status;
    },
  });
  const unknown = useMutation({ mutationFn: () => api.get("/nope") });
  const [typed, setTyped] = useState("");
  const [date, setDate] = useState(todayISO());

  return (
    <main style={{ padding: 24, display: "flex", flexDirection: "column", gap: 16 }}>
      <h1 style={{ fontFamily: "var(--inter)" }}>GoGym</h1>
      <p>
        وضعیت API: <b>{health.isPending ? "…" : health.isError ? "در دسترس نیست" : health.data}</b>
      </p>
      <button type="button" onClick={() => unknown.mutate()}>
        درخواست به مسیر ناموجود
      </button>
      {unknown.isError && <p data-testid="api-error">{unknown.error.message}</p>}
      <label className={f.fl}>
        <span className={f.label}>عدد آزمایشی</span>
        <input
          className={f.in}
          inputMode="decimal"
          value={typed}
          onChange={(e) => setTyped(e.target.value)}
        />
      </label>
      <p>
        مقدار خوانده‌شده: <b data-testid="parsed">{String(parseNumber(typed))}</b>
      </p>
      <JalaliDateField label="تاریخ آزمایشی" value={date} onChange={setDate} />
      <p dir="ltr" data-testid="iso">
        {date}
      </p>
    </main>
  );
}
