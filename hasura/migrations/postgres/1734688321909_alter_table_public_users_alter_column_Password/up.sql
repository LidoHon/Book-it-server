ALTER TABLE "public"."users" ALTER COLUMN "Password" TYPE varchar;
alter table "public"."users" rename column "Password" to "password";
