alter table "public"."wishlist" add column "created_at" timestamptz
 null default now();
