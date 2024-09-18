alter table "public"."users" drop constraint "email_format_check";
alter table "public"."users" add constraint "email_format_check" check (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9]+\.[A-Za-z]{2,}$'::text);
