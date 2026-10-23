
create or replace function public.set_primary_phone(p_phone_id uuid)
returns void
language plpgsql
security definer
set search_path = public
as $$
begin
  -- Verify caller owns this phone number; raises if not found/unauthorized.
  if not exists (
    select 1 from public.user_phone_numbers
    where id = p_phone_id and user_id = auth.uid()
  ) then
    raise exception 'phone_not_found';
  end if;

  -- Clear existing primary, then set the new one — both in this transaction.
  update public.user_phone_numbers
    set is_primary = false
    where user_id = auth.uid() and is_primary = true;

  update public.user_phone_numbers
    set is_primary = true
    where id = p_phone_id;
end;
$$;
;
