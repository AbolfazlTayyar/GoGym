-- Visit-by-visit tracking for athletes who pay per 12 sessions (یک روز در میان). Opt-in per plan: a plan
-- without a session_plan row is untracked, so coaches who never use it see no change.
CREATE TABLE session_plan (
    plan_id    UUID PRIMARY KEY REFERENCES plan (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One visit per day per plan; references session_plan so turning tracking off removes the visits with it.
CREATE TABLE attendance (
    plan_id    UUID NOT NULL REFERENCES session_plan (plan_id) ON DELETE CASCADE,
    date       DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (plan_id, date)
);
