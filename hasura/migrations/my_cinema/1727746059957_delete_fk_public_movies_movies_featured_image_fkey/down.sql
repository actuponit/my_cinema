alter table "public"."movies"
  add constraint "movies_featured_image_fkey"
  foreign key ("featured_image")
  references "public"."movie_thumbnails"
  ("image_url") on update set null on delete set null;
