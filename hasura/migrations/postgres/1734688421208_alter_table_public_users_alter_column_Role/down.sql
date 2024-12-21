alter table "public"."users" rename column "rol" to "Role";
ALTER TABLE "public"."users" ALTER COLUMN "Role" drop default;
ALTER TABLE "public"."users" ALTER COLUMN "Role" TYPE text;
