alter table "public"."schedules"
  add constraint "schedules_hall_fkey"
  foreign key ("hall")
  references "public"."halls"
  ("name") on update restrict on delete restrict;
