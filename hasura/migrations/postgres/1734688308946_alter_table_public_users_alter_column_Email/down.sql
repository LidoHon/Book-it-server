alter table "public"."users" rename column "email" to "Email";
ALTER TABLE "public"."users" ALTER COLUMN "Email" TYPE text;
