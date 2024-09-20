comment on column "public"."movie_thumbnails"."image_path" is E'Different thumbnails of a movie';
alter table "public"."movie_thumbnails" alter column "image_path" drop not null;
alter table "public"."movie_thumbnails" add column "image_path" text;
