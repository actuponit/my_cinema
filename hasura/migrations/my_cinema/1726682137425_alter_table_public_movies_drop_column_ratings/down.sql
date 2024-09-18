comment on column "public"."movies"."ratings" is E'the movie collections';
alter table "public"."movies" alter column "ratings" set default '0'::numeric;
alter table "public"."movies" alter column "ratings" drop not null;
alter table "public"."movies" add column "ratings" numeric;
