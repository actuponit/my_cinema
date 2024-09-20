comment on column "public"."movie_thumbnails"."featured" is E'Different thumbnails of a movie';
alter table "public"."movie_thumbnails" alter column "featured" set default false;
alter table "public"."movie_thumbnails" alter column "featured" drop not null;
alter table "public"."movie_thumbnails" add column "featured" bool;
