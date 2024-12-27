alter table "public"."rentedBooks" drop constraint "rentedBooks_book_id_fkey",
  add constraint "rentedBooks_book_id_fkey"
  foreign key ("book_id")
  references "public"."books"
  ("id") on update restrict on delete cascade;
