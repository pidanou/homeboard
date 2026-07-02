-- Older client/server code could save birthday events with non-midnight UTC
-- boundaries or multi-day spans. Backfill them to a single UTC-midnight day,
-- matching the normalization the service layer now enforces on every write.
UPDATE events
SET start_at = date_trunc('day', start_at),
    end_at = date_trunc('day', start_at) + interval '1 day',
    all_day = true
WHERE birthday_of IS NOT NULL AND birthday_of != '';
