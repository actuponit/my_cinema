alter table "public"."ratings"
  add constraint "ratings_user_fkey"
  foreign key ("user")
  references "public"."users"
  ("id") on update cascade on delete cascade;
