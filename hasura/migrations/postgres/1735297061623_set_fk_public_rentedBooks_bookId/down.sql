alter table "public"."rentedBooks" drop constraint "rentedBooks_bookId_fkey",
  add constraint "rentedBooks_book_id_fkey"
  foreign key ("bookId")
  references "public"."books"
  ("id") on update restrict on delete cascade;
