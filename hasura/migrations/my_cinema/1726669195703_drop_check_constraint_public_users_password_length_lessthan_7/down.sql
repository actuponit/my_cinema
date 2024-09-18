alter table "public"."users" add constraint "password_length_lessthan_7" check (CHECK (password ~ '^.{7,}$'::text));
