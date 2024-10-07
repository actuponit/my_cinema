alter table "public"."schedules" alter column "id" set default nextval('schedules_id_seq'::regclass);
