alter table "public"."ratings" alter column "created_at" set default now();
alter table "public"."ratings" alter column "created_at" drop not null;
alter table "public"."ratings" add column "created_at" timestamptz;
