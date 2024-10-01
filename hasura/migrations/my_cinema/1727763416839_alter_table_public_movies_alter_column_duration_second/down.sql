alter table "public"."movies" rename column "duration" to "duration_second";
comment on column "public"."movies"."duration_second" is NULL;
