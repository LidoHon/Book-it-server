alter table "public"."rentedBooks" drop constraint "rentedBooks_user_id_fkey",
  add constraint "rentedBooks_userId_fkey"
  foreign key ("userId")
  references "public"."users"
  ("id") on update restrict on delete cascade;
