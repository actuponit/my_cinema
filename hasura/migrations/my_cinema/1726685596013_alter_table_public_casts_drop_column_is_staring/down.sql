comment on column "public"."casts"."is_staring" is E'The movie crew';
alter table "public"."casts" alter column "is_staring" set default false;
alter table "public"."casts" alter column "is_staring" drop not null;
alter table "public"."casts" add column "is_staring" bool;
