alter table "public"."bookmarks"
  add constraint "bookmarks_schedule_fkey"
  foreign key ("movie_id")
  references "public"."schedules"
  ("id") on update cascade on delete cascade;
