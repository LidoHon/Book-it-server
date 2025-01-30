CREATE TABLE "public"."payments" ("id" serial NOT NULL, "rent_id" integer NOT NULL, "tx_ref" varchar NOT NULL, "checkout_url" varchar NOT NULL, "amount" integer NOT NULL, "currency" varchar NOT NULL, "payment_method" varchar, "status" varchar NOT NULL DEFAULT 'pending', "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), "user_id" integer, PRIMARY KEY ("id") , FOREIGN KEY ("rent_id") REFERENCES "public"."rentedBooks"("id") ON UPDATE no action ON DELETE cascade, FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON UPDATE no action ON DELETE cascade, UNIQUE ("tx_ref"));
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
CREATE TRIGGER "set_public_payments_updated_at"
BEFORE UPDATE ON "public"."payments"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_payments_updated_at" ON "public"."payments"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
