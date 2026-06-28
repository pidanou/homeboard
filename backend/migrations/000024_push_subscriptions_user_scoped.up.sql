DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'push_subscriptions' AND column_name = 'family_id'
  ) THEN
    ALTER TABLE push_subscriptions DROP COLUMN family_id;
  END IF;
END $$;
