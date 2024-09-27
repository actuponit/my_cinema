alter table "public"."schedules" alter column "start_time" drop not null;
alter table "public"."schedules" add column "start_time" timetz;
