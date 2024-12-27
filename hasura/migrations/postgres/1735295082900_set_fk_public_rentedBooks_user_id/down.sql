alter table "public"."rentedBooks" drop constraint "rentedBooks_user_id_fkey",
  add constraint "rentedBooks_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update restrict on delete restrict;
