alter table "public"."schedules" add constraint "start_in_future" check (start_time>NOW());
