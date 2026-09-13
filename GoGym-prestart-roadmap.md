# Pre-Start Roadmap: Gym Trainee Management App (GoGym)

**Goals:**

1. Expand Go expertise
2. Learn infrastructure/deployment concerns
3. Learn the end-to-end software development lifecycle
4. Design and ship a real app

**Context:** Full PWA (API + frontend + auth UI), Go level: basics, wants real backend and software engineering depth. The goal is to design, build, deploy, and ship the application from scratch. Deployment target open (recommendation included below).

---

## 1. Nail down the MVP scope

Write a one-page spec: who are the users (coaches, trainees, admins?), and what's the smallest useful feature set — e.g. coach creates trainees, assigns workout plans, logs progress. Cut anything that isn't essential for v1 (payments, notifications, analytics can wait).

## 2. Sketch the architecture on paper

Before writing code, decide on layers: handlers (HTTP) → services (business logic) → repositories (data access) → models. This is the classic Go "clean architecture" shape and it's the best way to actually learn idiomatic backend structure rather than picking it up ad hoc.

## 3. Choose your stack deliberately

Pick one option per concern and stick with it:

- **Router:** `net/http` with Go 1.22+ routing, or `chi`
- **DB:** PostgreSQL (standard choice)
- **DB access:** `sqlc` for type-safe SQL, or GORM if you want an ORM
- **Auth:** JWT + bcrypt
- **Migrations:** `goose` or `golang-migrate`
- **Frontend:** a separate SPA (React), or Go templates + htmx to stay Go-focused

## 4. Set up project structure and tooling

Use the standard Go project layout (`cmd/`, `internal/`, `pkg/` if needed). Set up `go.mod`, a Makefile for common tasks, `golangci-lint`, and a `.env`-based config loader from the start — retrofitting these later is painful.

## 5. Design the data model early

Draw an ER diagram for coaches, trainees, plans, sessions, progress logs before touching code. Write your first migration files. Getting the schema roughly right early saves a lot of rework.

## 6. Containerize your dev environment

Write a `docker-compose.yml` with your Go app + Postgres from day one, even for local dev. This forces you to think about config, env vars, and networking early — which directly feeds the infra-learning goal.

## 7. Decide on a deployment target for v1

Start simple: a VPS (Hetzner or DigitalOcean, ~$5-10/mo) running Docker Compose. It teaches you real deployment concerns (reverse proxy, TLS via Caddy or nginx, systemd/Docker restart policies, backups) without Kubernetes' complexity. Save k8s for a v2 "infra learning" pass once the app is stable — it's much more useful to learn on something that already works.

## 8. Set up CI from the start

A basic GitHub Actions pipeline that runs `go vet`, `go test`, and `golangci-lint` on every push. Add automated Docker image builds once you're deploying. This is a low-cost, high-value habit to build early rather than bolt on later.

---

## Notes on the three goals

- **Goal 1 (deepen Go):** served most by step 2 — forcing yourself into a layered architecture instead of one giant `main.go` is where the real learning happens, more than any specific library choice.
- **Goal 2 (infra):** resist the urge to jump straight to Kubernetes. A VPS + Docker Compose + Caddy teaches you 80% of the infra concepts (networking, TLS, process management, backups, monitoring) with 20% of the complexity. You can migrate the same app to k8s later as a dedicated learning exercise once it's boring and stable.
- **Goal 3 (ship it):** the MVP-scope step is the one people skip and regret. Gym trainee management can balloon (scheduling, payments, messaging, progress photos...). Write down what's *out* of scope for v1 as explicitly as what's in.

## Frontend decision to make deliberately

Since you want deep Go, consider **Go templates + htmx** instead of a separate React app — it keeps you writing Go for longer and avoids context-switching into a JS build toolchain, while still giving you a real interactive UI.

If you'd rather also build frontend skills separately, a **React SPA talking to your Go API** is the more "standard" split — just know it roughly doubles your surface area to learn.

## Practical first move

Before any code, write the one-page spec (users, core entities, MVP feature list) and the ER diagram. Everything else — folder structure, routing, docker-compose — flows naturally once that's nailed down.
