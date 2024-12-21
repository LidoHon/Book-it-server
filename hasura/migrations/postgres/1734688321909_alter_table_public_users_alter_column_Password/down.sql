alter table "public"."users" rename column "password" to "Password";
ALTER TABLE "public"."users" ALTER COLUMN "Password" TYPE text;
