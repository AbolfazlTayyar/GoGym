import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

// The browser sees one origin in dev, as behind Caddy in production, so the auth cookie needs no CORS.
const apiProxy = {
  "/api": "http://localhost:8080",
  "/healthz": "http://localhost:8080",
};

export default defineConfig({
  plugins: [react()],
  server: { proxy: apiProxy },
  preview: { proxy: apiProxy },
});
