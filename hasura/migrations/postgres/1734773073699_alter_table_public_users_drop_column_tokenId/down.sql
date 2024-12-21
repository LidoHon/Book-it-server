alter table "public"."users" alter column "tokenId" drop not null;
alter table "public"."users" add column "tokenId" varchar;
