import { useQuery } from "@tanstack/react-query";

// Proves the dev proxy reaches the API; the sign-in screen replaces it.
export function Placeholder() {
  const health = useQuery({
    queryKey: ["healthz"],
    queryFn: async () => {
      const res = await fetch("/healthz");
      const body: { status: string } = await res.json();
      return body.status;
    },
  });

  return (
    <main style={{ fontFamily: '"Vazirmatn Variable", sans-serif', padding: 24 }}>
      <h1 style={{ fontFamily: '"Inter Variable", sans-serif' }}>GoGym</h1>
      <p>
        وضعیت API: <b>{health.isPending ? "…" : health.isError ? "در دسترس نیست" : health.data}</b>
      </p>
    </main>
  );
}
