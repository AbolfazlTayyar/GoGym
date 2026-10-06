-- A plan is at most one training week: 7 days. A unique, bounded order_index
-- caps the count structurally, so concurrent inserts can't race past it the
-- way an application-side count check could.
ALTER TABLE day
    ADD CONSTRAINT day_plan_id_order_index_key UNIQUE (plan_id, order_index),
    ADD CONSTRAINT day_order_index_range CHECK (order_index BETWEEN 0 AND 6);
