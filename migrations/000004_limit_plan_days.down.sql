ALTER TABLE day
    DROP CONSTRAINT IF EXISTS day_order_index_range,
    DROP CONSTRAINT IF EXISTS day_plan_id_order_index_key;
