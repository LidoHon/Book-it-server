alter table "public"."users" rename column "phone" to "Phone";
ALTER TABLE "public"."users" ALTER COLUMN "Phone" TYPE text;
