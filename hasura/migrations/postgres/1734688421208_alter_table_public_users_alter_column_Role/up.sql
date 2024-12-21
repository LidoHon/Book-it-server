ALTER TABLE "public"."users" ALTER COLUMN "Role" TYPE varchar;
alter table "public"."users" alter column "Role" set default 'user';
alter table "public"."users" rename column "Role" to "rol";
