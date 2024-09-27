alter table "public"."schedules"
  add constraint "schedules_format_fkey"
  foreign key ("format")
  references "public"."formats"
  ("name") on update restrict on delete restrict;
