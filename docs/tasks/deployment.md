# GoGym — Deployment tasks (v1)

Getting the API onto a real server and keeping it running. Needs the backend API from [backend.md](backend.md). Task format is explained in [README.md](README.md).

---

### VPS deployment & production migrations

⬜ **Not started**

**Description:** Nothing runs anywhere but local dev and CI today — there is no deploy target and no defined way to apply a migration against a real production database. This task exists because the *Backup & restore drill* task below assumes a running deployment target and a migration process already exist; right now neither does.

**Prompt:**
```
Stand up the v1 deployment target per docs/archive/prestart-roadmap.md step 7: a VPS
(Hetzner/DigitalOcean or similar) running the existing docker-compose.yml (api + db),
behind a reverse proxy (Caddy is the simplest TLS option) terminating HTTPS. Don't reach
for Kubernetes.

Run Caddy as a Compose service with a fixed IP (a network with a pinned subnet and
`ipv4_address` on caddy), and set TRUSTED_PROXIES to exactly that IP (docs/adr/0024).
Container IPs change across restarts unless pinned, and an empty TRUSTED_PROXIES puts every
client in Caddy's auth rate-limit bucket. Never widen it to 0.0.0.0/0 — that brings
back X-Forwarded-For spoofing.

On the VPS, don't publish the api (8080) or db (5432) ports — docker-compose.yml publishes
both on all interfaces for local dev. Only Caddy's 80/443 face the internet; a published
5432 exposes Postgres, and a published 8080 lets callers skip Caddy. Use a prod override
file or bind them to 127.0.0.1 if you need them for SSH tunnels — say which.

Define and document how migrations reach production — this app has no auto-migrate on
container startup today (internal/db explicitly disables GORM AutoMigrate; migrations/
is the only source of truth, applied via `make migrate-up` wrapping golang-migrate CLI).
Pick one deliberately and tell me why:
- A manual step: SSH to the VPS, run `make migrate-up` (or the raw `migrate` binary)
  against the prod DSN before restarting the api container on each deploy.
- An automated step: a one-shot init container/job in the compose stack that runs
  `migrate ... up` and exits before the api service starts (compose depends_on +
  a healthcheck-style gate, or a deploy script step) — safer against "forgot to migrate"
  but needs care that it never runs concurrently with a second deploy.

Whichever you pick, the DSN must come from the same internal/config-driven env vars used
locally — no separate prod-only connection logic. Add a short runbook (a doc or a section
in this file) covering: how to deploy a new image, how/when migrations run relative to
that, and how to roll back a bad migration (down migration + previous image tag).

Per docs/archive/prestart-roadmap.md step 10, this is also the natural point to wire basic
CD (a GitHub Actions job that builds and pushes the image, then SSHs to the VPS to pull
and restart — no need for anything fancier yet) — call out explicitly if you're deferring
that to a separate task instead of doing it here.
```

**Checkpoint:** A real migration (start with the existing `000001`/`000002` files) has actually been applied against the VPS's Postgres, not just described — confirm via `\dt`/`\d athlete` over SSH. Deploying a trivial code change (e.g. a log line) end-to-end once, following only the written runbook, succeeds without undocumented manual steps. Rolling back one migration via the down file has been exercised at least once, not just assumed to work. A request from your own machine logs your real public IP as `client_ip` (not Caddy's), and the same request with `-H "X-Forwarded-For: 1.2.3.4"` still logs your real IP. From outside the VPS, ports 8080 and 5432 can't be reached.

**Expected output:**
```
$ curl -s https://<your-domain>/healthz
{"status":"ok"}

$ nc -zv -w3 <vps-ip> 5432
nc: connect to <vps-ip> port 5432 (tcp) failed: Connection refused

$ ssh vps 'docker compose exec db psql -U gogym -c "\dt"'
 public | athlete              | table | gogym
 public | block                | table | gogym
 ...
 public | schema_migrations    | table | gogym

$ ssh vps 'docker compose exec db psql -U gogym -c "SELECT version, dirty FROM schema_migrations"'
 version | dirty
---------+-------
       2 | f
```

**In short:** the app runs on a real server over HTTPS, and there's a written, tested way to ship a new version and its migrations — and to undo one.

---

### Backup & restore drill **[A4]**

⬜ **Not started**

**Description:** Not a feature — the point at which losing the database stops being an inconvenience and starts being the end of the business. An untested backup is not a backup.

**Prompt:**
```
Set up database backups for the deployment target from docs/archive/prestart-roadmap.md step 7
(VPS + Docker Compose): a scheduled pg_dump of the Postgres volume, retained for a
sensible window, stored somewhere that is NOT the same VPS — a lost disk shouldn't take
the backups with it. Keep it as simple as the rest of the stack; a cron job and an
object-storage bucket is enough, this doesn't need a backup tool.

Then actually perform a restore drill and write down what happened: take a dump, restore
it into a throwaway Postgres (locally or a second container), and confirm the data is
really there. Document the restore procedure as a short runbook — the steps, in order,
that you'd follow at 2am with a dead database. Note how long the restore took.

Don't claim this task is done on the strength of the backup job running. It's done when
a restore has succeeded and the runbook has been followed once, start to finish.
```

**Checkpoint:** A backup exists off the VPS. A restore from that backup into an empty Postgres has actually been run by you — not described — and the restored database serves a working `/healthz` and returns real athlete rows. The runbook exists and someone who isn't you could follow it.

**Expected output:**
```
$ <list the off-VPS bucket>
gogym-2026-10-02T03-00.sql.gz
gogym-2026-10-03T03-00.sql.gz

$ gunzip -c gogym-2026-10-03T03-00.sql.gz | docker compose -f restore.yml exec -T db psql -U gogym
...
$ curl -s localhost:8081/healthz
{"status":"ok"}
$ docker compose -f restore.yml exec db psql -U gogym -c "SELECT count(*) FROM athlete"
 count
-------
    37
```

**In short:** the database is copied off the server every day, and you've actually brought it back from a copy once, so you know it works.

---

### Uptime monitoring & alerting

⬜ **Not started**

**Description:** The `api` container has a healthcheck, but Compose only turns it into a status label — nothing restarts the app and nobody is told when it goes down. A check nobody watches catches nothing. This task makes an outage reach a person, and makes a crash recover on its own.

**Prompt:**
```
Make production failures visible and self-healing where restarting actually helps. Needs
the VPS deployment and backup tasks above to be done: there must be a public URL and a
backup job to watch.

1. Add `restart: unless-stopped` to the api and db services in docker-compose.yml, so a
   crashed process or a rebooted VPS comes back without anyone SSHing in. Don't restart
   on `unhealthy` (no autoheal container): an unhealthy api almost always means the
   database is down, and restarting the api doesn't fix that — it needs a person.
   To test it, kill the process from the host (`sudo kill -9 <container's host PID>`):
   `docker kill`/`docker stop` count as a manual stop, so the policy never restarts them.

2. Set up an external uptime monitor (UptimeRobot, Better Stack, or Uptime Kuma hosted
   somewhere other than the VPS — pick one and tell me why) that GETs
   https://<your-domain>/healthz every 1–5 minutes and alerts me (email or Telegram) when
   it fails and again when it recovers. It must run outside the VPS: nothing on the box
   can report the box itself, its network, Caddy or an expired TLS certificate being down.

3. Give the backup job from the Backup & restore drill task a heartbeat (dead-man's switch,
   e.g. Healthchecks.io): the job pings a URL after a successful upload, and I'm alerted
   when the ping doesn't arrive on schedule. A backup that silently stops is the failure
   nobody notices until the restore.

Don't add a metrics/dashboard stack (Prometheus, Grafana) here — alerting on "down" and
"backup missed" is the whole scope.
```

**Checkpoint:** Stop `db` on the VPS and the alert reaches you within the monitor's interval plus a few minutes; start it again and the recovery notice arrives. Kill the api process from the host (not with `docker kill`) and the container comes back on its own. The 503 body during the outage has a fixed reason with no hostnames or IPs, and the real error is in the api logs. Skip one scheduled backup (or point its heartbeat at the wrong URL) and the missed-heartbeat alert arrives.

**Expected output:**
```
$ ssh vps 'docker compose stop db'
$ curl -s https://<your-domain>/healthz
{"status":"unavailable","reason":"database unreachable"}
# ...alert arrives: "gogym /healthz is DOWN (503)"

$ ssh vps 'docker compose start db'
# ...alert arrives: "gogym /healthz is UP"

$ ssh vps 'sudo kill -9 $(docker inspect -f "{{.State.Pid}}" $(docker compose ps -q api))'
$ ssh vps 'sleep 15 && docker inspect -f "{{.RestartCount}} {{.State.Health.Status}}" $(docker compose ps -q api)'
1 healthy
```

**In short:** if the app or its backups stop working, you hear about it from your phone instead of from a coach, and a crash fixes itself.

---

### Hide /healthz behind a secret path

⬜ **Not started**

**Description:** `/healthz` is unauthenticated and pings the database on every request, so a flood of it competes with real traffic for the connection pool. Nothing public needs it except the uptime monitor: the container healthcheck calls it from inside the container, and after the VPS task the api port isn't published. This task makes Caddy refuse the plain path and serve the check only on a hard-to-guess one that the monitor uses. Needs the VPS deployment and uptime monitoring tasks above.

**Prompt:**
```
Stop serving /healthz to the public through Caddy, without losing the external uptime
monitor's database check.

- Generate a random path segment (e.g. `openssl rand -hex 16`) and keep it in the VPS's
  env as HEALTH_PATH, passed to the caddy service. Never commit the real value; if you
  add a placeholder to an example env file, say which.
- In the Caddyfile, use mutually exclusive handle blocks: plain /healthz answers 404,
  /{$HEALTH_PATH} is rewritten to /healthz and proxied to the api, everything else is
  proxied as before:

    your-domain.com {
        handle /healthz {
            respond 404
        }
        handle /{$HEALTH_PATH} {
            rewrite * /healthz
            reverse_proxy api:8080
        }
        handle {
            reverse_proxy api:8080
        }
    }

  The api still sees /healthz, so the healthy-probe log skip and the 503 logging keep
  working unchanged. Don't change the api or the container healthcheck: neither goes
  through Caddy.
- An unset HEALTH_PATH makes the second block `handle /`, which serves the health check
  at the site root. Make the deploy fail when it's empty (e.g. `${HEALTH_PATH:?}` in the
  compose environment) rather than relying on someone noticing.
- Point the uptime monitor at https://<your-domain>/<HEALTH_PATH>. Keep that URL off any
  public status page the monitor offers.
- Add to the deploy runbook how to rotate the secret: new value in the env, restart
  caddy, update the monitor URL. A leaked value exposes only the health check, so
  rotation is cheap, not an emergency.
```

**Checkpoint:** From outside the VPS, plain `/healthz` and its variants all get Caddy's 404, not the api's response: `/healthz`, `/healthz/`, `/HEALTHZ`, `//healthz`, and `curl --path-as-is https://<your-domain>/api/../healthz`. The secret path returns `{"status":"ok"}`. Stop `db`: the secret path returns the 503 and the uptime monitor alerts, so the database check survived the change. `docker compose ps` still shows the api `healthy`. With `HEALTH_PATH` unset, `docker compose up` refuses to start rather than serving the check at `/`.

**Expected output:**
```
$ curl -s -o /dev/null -w "%{http_code}\n" https://<your-domain>/healthz
404

$ curl -s https://<your-domain>/$HEALTH_PATH
{"status":"ok"}

$ ssh vps 'docker compose stop db' && curl -s https://<your-domain>/$HEALTH_PATH
{"status":"unavailable","reason":"database unreachable"}
# ...uptime monitor alert arrives
```

**In short:** strangers can't reach the health check, and your uptime monitor still can, so it still notices when the database goes down.
