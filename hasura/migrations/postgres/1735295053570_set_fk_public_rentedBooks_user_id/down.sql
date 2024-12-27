alter table "public"."rentedBooks" drop constraint "rentedBooks_user_id_fkey",
  add constraint "RentedBooks_user_id_fkey"
  foreign key ("book_id")
  references "public"."books"
  ("id") on update restrict on delete restrict;
