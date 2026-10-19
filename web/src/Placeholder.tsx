import { useMutation, useQuery } from "@tanstack/react-query";
import { api } from "./api/client";

// Proves the dev proxy reaches the API and that API errors read in Persian; the sign-in screen replaces it.
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

  return (
    <main style={{ padding: 24 }}>
      <h1 style={{ fontFamily: "var(--inter)" }}>GoGym</h1>
      <p>
        وضعیت API: <b>{health.isPending ? "…" : health.isError ? "در دسترس نیست" : health.data}</b>
      </p>
      <button type="button" onClick={() => unknown.mutate()} style={{ marginTop: 16 }}>
        درخواست به مسیر ناموجود
      </button>
      {unknown.isError && <p data-testid="api-error">{unknown.error.message}</p>}
    </main>
  );
}
