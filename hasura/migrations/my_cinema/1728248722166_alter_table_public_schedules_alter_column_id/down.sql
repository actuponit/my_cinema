alter table "public"."schedules" alter column "id" set default pg_event_trigger_table_rewrite_reason();
ALTER TABLE "public"."schedules" ALTER COLUMN "id" TYPE integer;
