# 0016 — The current plan is the latest one that has started

- Status: accepted
- Date: 2026-10-06
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The athlete's plan list highlights the current plan and shows the others as history
(spec.md, MVP feature 3). The spec doesn't say which plan counts as current. A plan has a
`start_date` and no end date or status, and a coach can write the next plan before it starts.

## Decision Drivers

- Writing next month's plan early shouldn't hide the plan the athlete is training on today.
- No schema change for an MVP list screen.
- The server computes `is_current` so every client agrees.

## Considered Options

1. The plan with the latest `start_date`, whether or not it has started
2. The plan with the latest `start_date` on or before today
3. An explicit `is_current` / status column the coach sets

## Decision Outcome

Chosen option: **2**. "Today" is the server's local date. When two plans start on the same day,
the later-created one is current. If every plan starts in the future, none is current.

### Consequences

- Good: a future-dated plan reads as upcoming, and it becomes current on its start date with no write.
- Good: works on the existing schema.
- Bad: "today" follows the server's time zone, not the coach's. A plan starting today becomes current
  at server midnight, which may be hours off from the coach's midnight.
- Bad: the coach can't keep an older plan current after a newer one starts. That needs option 3.
