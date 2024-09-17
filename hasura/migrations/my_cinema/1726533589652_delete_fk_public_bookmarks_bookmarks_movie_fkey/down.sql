alter table "public"."bookmarks"
  add constraint "bookmarks_movie_fkey"
  foreign key ("schedule")
  references "public"."movies"
  ("id") on update cascade on delete cascade;
