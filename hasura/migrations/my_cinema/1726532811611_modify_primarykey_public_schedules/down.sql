alter table "public"."schedules"
    add constraint "schedules_pkey"
    primary key ("movie", "start_time");
