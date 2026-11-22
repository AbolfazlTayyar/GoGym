# 0030 — React single-page app, served by Caddy on the API's origin

- Status: accepted
- Date: 2026-10-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The API's six MVP features are built, and the design is chosen ([0029](0029-blush-theme-day-buttons-home.md)).
The frontend needs a stack. The [prestart roadmap](../archive/prestart-roadmap.md) left two options
open: a React SPA, or Go templates + htmx to keep the work in Go. Since then the priority has
changed. A good frontend matters more than learning, and learning frontend isn't a goal for now.
Offline use, SEO and more coaches aren't needed yet, but may be soon.

## Decision Drivers

- Mobile first: the coach does everything on a phone, building whole plans included
  ([stakeholder brief](../stakeholder-brief.md) §2).
- The builder edits a whole day in the browser before one save: steppers, presets, supersets,
  reordering and movement search ([concept-8-1](../design/concept-8-1-home-and-plans.png)).
- The `/api/v1` JSON API, with its envelope and Swagger docs, already exists.
- Persian RTL, Jalali dates and Persian digits, and one measurement chart.
- A PWA on one phone, online only today. Reading plans offline may be wanted later.
- SEO may matter later, for a landing page or share-link previews, not for the logged-in app.

## Considered Options

1. React + Vite + TypeScript SPA
2. Vue 3 + Vite SPA
3. Svelte 5 + Vite SPA
4. Go templates (or templ) + htmx + Alpine.js
5. Next.js (or Nuxt / SvelteKit) with server rendering

## Decision Outcome

Chosen option: **1**, a static build served by Caddy on the same origin as the API: `/api/*` and
`/healthz` go to the Go API, everything else is the built files with an `index.html` fallback.
React Router v7 handles routes and TanStack Query handles data fetching. Every API call goes
through one `src/api/` module.

React has the most ready-made parts for these screens (bottom sheets, Jalali date pickers,
reordering). It also has the most documentation and examples, which matters more than ease of
learning now that learning frontend isn't a goal. Vue (2) was a close second only on ease of
learning. Svelte (3) has the fewest Persian/Jalali components, and Svelte 5 made much of the
existing help outdated. htmx (4) leaves the JSON API unused by the UI and needs a second set of HTML
handlers. Its builder needs either a server round-trip per tap or Alpine.js holding a client-side
draft, which means two UI models. It also can't work offline. Next.js (5) adds a Node server
to run in production. Server rendering doesn't help an app that is only usable after login.

Serving from the API's origin means no CORS, and a `SameSite=Strict` cookie needs no CSRF token
(see *Browser auth via httpOnly cookie* in [tasks/backend.md](../tasks/backend.md)). The files
are served by Caddy, not embedded in the Go binary. Embedding would need an SPA fallback that
clashes with the API's enveloped `NoRoute` 404.

The possible future needs stay cheap without building anything now:
- **Offline reading:** a service worker (`vite-plugin-pwa`) and TanStack Query's IndexedDB
  persister. That's mostly configuration, and nothing about htmx would allow it.
- **Offline editing:** the cost is mostly on the backend (a queue of saves to replay, and
  conflict rules). The whole-day save in *Save a day* can be replayed safely.
- **SEO:** React Router v7's framework mode can pre-render chosen pages to static HTML with
  `ssr: false`, so still with no Node server, or move to SSR without a rewrite.
- **More coaches:** static files cost the server nothing, and all the load stays on the API.

### Consequences

- Good: the API stays the only contract, and Swagger documents what the UI uses.
- Good: the builder's draft lives in browser state and is saved in one request.
- Good: offline reading and pre-rendered pages can be added later without a rewrite.
- Bad: a second toolchain (Node, npm, TypeScript) in the repo and in CI.
- Bad: API types are hand-written in TypeScript. `swag` v1 emits Swagger 2.0, and
  `openapi-typescript` needs OpenAPI 3, so the types can drift from the Go structs without anything
  noticing.
- Bad: release order matters. Caddy serves a new frontend as soon as it's deployed, so an
  endpoint the frontend calls must be deployed first.
