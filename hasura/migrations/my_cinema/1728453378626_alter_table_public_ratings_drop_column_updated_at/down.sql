alter table "public"."ratings" alter column "updated_at" set default now();
alter table "public"."ratings" alter column "updated_at" drop not null;
alter table "public"."ratings" add column "updated_at" timestamptz;
