alter table "public"."users" alter column "rating" set default '0'::numeric;
alter table "public"."users" alter column "rating" drop not null;
alter table "public"."users" add column "rating" numeric;
