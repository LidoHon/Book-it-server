alter table "public"."rentedBooks" alter column "payment_status" set default ''pending'::character varying';
alter table "public"."rentedBooks" alter column "payment_status" drop not null;
alter table "public"."rentedBooks" add column "payment_status" varchar;
