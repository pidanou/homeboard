ALTER TABLE lists ADD COLUMN position DOUBLE PRECISION NOT NULL DEFAULT 0;

WITH ordered AS (
    SELECT id, row_number() OVER (PARTITION BY family_id ORDER BY created_at) - 1 AS rn
    FROM lists
)
UPDATE lists SET position = ordered.rn
FROM ordered
WHERE lists.id = ordered.id;
