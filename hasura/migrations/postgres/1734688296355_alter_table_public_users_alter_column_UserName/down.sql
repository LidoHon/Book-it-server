alter table "public"."users" rename column "username" to "UserName";
ALTER TABLE "public"."users" ALTER COLUMN "UserName" TYPE text;
