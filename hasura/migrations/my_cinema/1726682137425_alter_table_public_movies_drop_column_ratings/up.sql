alter table "public"."movies" add column "ratings" numeric;
alter table "public"."movies" alter column "ratings" set default '0'::numeric;
alter table "public"."movies" alter column "ratings" drop not null;
