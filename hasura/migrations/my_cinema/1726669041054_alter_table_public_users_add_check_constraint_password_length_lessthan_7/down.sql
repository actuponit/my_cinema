alter table "public"."users" drop constraint "password_length_lessthan_7";
alter table "public"."users" add constraint "password_length_lessthan_7" check (CHECK (char_length(password) > 6));
