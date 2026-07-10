DROP INDEX events_subscription_external_uid_uidx;

ALTER TABLE events
  DROP COLUMN external_uid,
  DROP COLUMN subscription_id;

DROP TABLE calendar_subscriptions;
