# GoGym — MVP Task List v1

Stack locked in for this list: **Go + Gin + GORM + PostgreSQL**, migrations via `golang-migrate`, JWT + bcrypt auth, module-first / clean-architecture-per-module layout (`internal/<module>/{handler,service,repository,model}.go`) per [prestart-roadmap.md](../archive/prestart-roadmap.md) step 2. API docs via **swaggo (Swagger UI)** for manual endpoint testing — see the *Swagger / OpenAPI docs setup* task. Frontend stack is a **mobile-first React + Vite + TypeScript SPA** (React Router, TanStack Query) in `web/`, built to static files and served by Caddy on the API's origin. It was chosen because the builder keeps a whole day in browser state, the JSON API already exists, and offline reading or pre-rendered pages can be added later without a rewrite. See [ADR 0030](../adr/0030-react-spa-same-origin.md).

Each task has: what it is, a plain description, a ready-to-paste prompt for Claude Code, and a checkpoint to verify it's actually done before moving on. Within a file, work top to bottom — later tasks assume earlier ones are merged.

> **Amendments folded in:** tasks marked **[A1]**–**[A4]** exist because the goal shifted from an internal tool to something other coaches pay for. They are grouped under those labels because each is close to free while there is no production data and expensive afterwards — do them in the order they appear, not last. The reasoning behind them is in [product-direction.md](../product-direction.md); the task text in each file is self-contained and is what you work from.

| File | What's in it | Order |
|---|---|---|
| [backend.md](backend.md) | Foundation, the core API features, the frontend stack decision, and deferred backend work | First |
| [deployment.md](deployment.md) | The v1 VPS deployment, backups, uptime alerting and hardening the public surface | Once the backend API is built |
| [frontend.md](frontend.md) | The React app, from scaffolding to every v1 screen and serving it from Caddy | Once the backend task each screen names is done |
