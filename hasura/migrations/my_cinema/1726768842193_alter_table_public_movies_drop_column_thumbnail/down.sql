comment on column "public"."movies"."thumbnail" is E'the movie collections';
alter table "public"."movies" alter column "thumbnail" drop not null;
alter table "public"."movies" add column "thumbnail" int4;
