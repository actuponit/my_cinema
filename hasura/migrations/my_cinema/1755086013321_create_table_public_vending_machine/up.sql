CREATE TABLE "public"."vending_machine" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "name" text NOT NULL, "location" text NOT NULL, "latitude" numeric NOT NULL, "longitude" numeric NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") );COMMENT ON TABLE "public"."vending_machine" IS E'Data about the vending machines';
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
CREATE TRIGGER "set_public_vending_machine_updated_at"
BEFORE UPDATE ON "public"."vending_machine"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_vending_machine_updated_at" ON "public"."vending_machine"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
CREATE EXTENSION IF NOT EXISTS pgcrypto;
