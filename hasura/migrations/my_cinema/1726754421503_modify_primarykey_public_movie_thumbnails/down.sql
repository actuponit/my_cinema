alter table "public"."movie_thumbnails" drop constraint "movie_thumbnails_pkey";
alter table "public"."movie_thumbnails"
    add constraint "movie_thumbnails_pkey"
    primary key ("image_path");
