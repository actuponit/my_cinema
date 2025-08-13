CREATE TABLE "public"."vending_machine_items" ("id" serial NOT NULL, "name" text NOT NULL, "price" numeric NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") );
CREATE OR REPLACE FUNCTION "public"."set_current_timestamp_updated_at"()
RETURNS TRIGGER AS $$
DECLARE
  _new record;
BEGIN
  _new := NEW;
  _new."updated_at" = NOW();
  RETURN _new;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER "set_public_vending_machine_items_updated_at"
BEFORE UPDATE ON "public"."vending_machine_items"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_vending_machine_items_updated_at" ON "public"."vending_machine_items"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
