---
name: run-on-phone
description: Launch the GoGym app (Postgres, Go API, Vite frontend) so it can be opened on a real phone over the LAN. Use when asked to run or start the app, or to test it on a phone.
model: haiku
---

# Run GoGym on a phone

The phone loads the Vite dev server over the LAN. Vite forwards `/api` and `/healthz` to the API on `:8080`, so the browser sees one origin and the auth cookie needs no CORS.

## Before starting

- `.env` must exist (`cp .env.example .env` if not), with `DB_HOST=localhost` and `APP_ENV=development`. Development is what turns off the cookie's `Secure` flag; without it, login can't stick over plain http.
- Run `npm ci` in `web/` if `web/node_modules` is missing.
- Ports 5173, 8080 and 5432 must be free. If a server is already listening on 5173 or 8080, check whether it's this app before starting another.

## Start it

From the repo root, start the database and apply migrations. `migrate` prints `no change` when the schema is current:

```sh
docker compose up -d db
docker compose up migrate
```

Build the API on the host, then start it and Vite as two long-running processes (each in its own terminal, or as a background task):

```sh
go build -o bin/api ./cmd/api
./bin/api
```

```sh
cd web && npm run dev -- --host
```

Don't use `docker compose up --build api` here: the image's `go mod download` can be refused by `proxy.golang.org`, while the host build uses the local module cache.

## Check it

Take the LAN address from Vite's `Network:` line, then replace `<lan-ip>` below:

```sh
curl -s http://<lan-ip>:5173/healthz
curl -s -o /dev/null -w '%{http_code}\n' http://<lan-ip>:5173/
curl -s -w ' [%{http_code}]\n' http://<lan-ip>:5173/api/v1/coaches/me
```

Expect `{"status":"ok"}`, then `200`, then an `unauthorized` envelope with `[401]`.

These run on the computer itself, so they don't prove the phone can reach it. Only opening the page on the phone does.

## Tell the user

- Give them `http://<lan-ip>:5173`. The phone must be on the same Wi-Fi, not a guest network or a VPN.
- They sign in with an existing coach account, as there is no signup screen. If they don't have one, `go run ./cmd/coachctl create -first-name <first> -last-name <last> -phone <phone>` makes it and reads the password from stdin.
- Frontend edits hot-reload on the phone. Go changes need the API rebuilt and restarted.
- Background tasks may stop when the Claude session closes, while the database container keeps running. Say so, and give them the two commands above that start the API and Vite.
