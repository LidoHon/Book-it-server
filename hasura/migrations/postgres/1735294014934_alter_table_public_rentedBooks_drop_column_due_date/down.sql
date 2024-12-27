alter table "public"."rentedBooks" alter column "due_date" drop not null;
alter table "public"."rentedBooks" add column "due_date" timestamptz;
