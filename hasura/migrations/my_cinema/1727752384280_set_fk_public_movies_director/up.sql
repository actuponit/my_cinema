alter table "public"."movies" drop constraint "movies_director_fkey",
  add constraint "movies_director_fkey"
  foreign key ("director")
  references "public"."casts"
  ("id") on update cascade on delete set null;
