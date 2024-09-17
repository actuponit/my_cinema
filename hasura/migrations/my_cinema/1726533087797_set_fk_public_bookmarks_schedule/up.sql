alter table "public"."bookmarks"
  add constraint "bookmarks_schedule_fkey"
  foreign key ("schedule")
  references "public"."schedules"
  ("id") on update cascade on delete cascade;
