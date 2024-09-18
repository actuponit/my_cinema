alter table "public"."users" add constraint "password_length_lessthan_7" check ((char_length(password) > 6));
