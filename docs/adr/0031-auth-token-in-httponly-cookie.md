# 0031 — The browser's auth token lives in an httpOnly cookie, with no CORS

- Status: accepted
- Date: 2026-10-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Login returned the JWT in the response body and `AuthMiddleware` read only the
`Authorization: Bearer` header. A browser client would have to keep that token where JavaScript
can read it, such as localStorage, where any injected script can steal it and use it from
anywhere until it expires. The frontend is a React SPA served on the API's origin
([0030](0030-react-spa-same-origin.md)), so the API also had to decide what CORS and CSRF
protection it needs.

## Decision Drivers

- A script injected into the page must not be able to read the token.
- Swagger UI's Authorize, curl and future non-browser clients keep working.
- No machinery the same-origin deployment doesn't need.

## Considered Options

1. Token in the login body, kept in localStorage, sent as a Bearer header
2. Token in an httpOnly cookie only, with the Bearer header still accepted
3. Option 2, plus the token kept in the login body
4. Option 2, plus CORS with an explicit origin list and a CSRF token

## Decision Outcome

Chosen option: **2**. Login sets `gogym_auth`: `HttpOnly`, `Secure` outside the development
environment (local dev and a phone on the LAN are plain http), `SameSite=Strict`,
`Path=/api/v1`, and `Max-Age` equal to `JWT_EXPIRY`. The login body is `data: null`.
`AuthMiddleware` takes the cookie, or else the Bearer header. `POST /api/v1/auth/logout`
clears the cookie, since page scripts can't; it needs no valid token, so an expired session can
still sign out.

Option 1 is what the cookie exists to avoid. Option 3 would hand the token back to page scripts
on every login; tools that want a token can read it from the `Set-Cookie` header (curl `-c`), and
Swagger UI on the same origin sends the cookie by itself. Option 4 solves a cross-origin problem
this deployment doesn't have: the browser only ever calls the API from its own origin, through
Vite's proxy in development and Caddy in production. So there is no CORS middleware at all, and no
other origin gets CORS headers. `SameSite=Strict` keeps the cookie off every request another site
starts, which is the CSRF protection; a token would add nothing on one origin. `Strict` rather
than `Lax` costs nothing here, because the app's HTML is static and needs no cookie; only the
SPA's own same-origin API calls carry it. `Path=/api/v1` keeps it off requests for the frontend's
files.

### Consequences

- Good: an injected script can call the API while the page is open, but can't take the token away.
- Good: no CORS configuration or CSRF token to maintain.
- Bad: serving the frontend from another site would need CORS with credentials and a CSRF
  defense; that is a new decision, not a config change.
- Bad: logout clears the cookie, but the token stays valid until it expires, as every JWT here
  does ([0010](0010-coach-jwt-auth.md)). The coach stays signed in for `JWT_EXPIRY`, so production
  should set it to weeks, not the example's 24 hours.
- Bad: a Bearer token copied out of a cookie keeps working for its lifetime, like before.
