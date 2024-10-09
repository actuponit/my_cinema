alter table "public"."ratings" add column "created_at" timestamptz
 null default now();
