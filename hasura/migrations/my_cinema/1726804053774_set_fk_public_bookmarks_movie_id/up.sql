alter table "public"."bookmarks"
  add constraint "bookmarks_movie_id_fkey"
  foreign key ("movie_id")
  references "public"."movies"
  ("id") on update cascade on delete cascade;
