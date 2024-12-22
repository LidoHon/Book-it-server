alter table "public"."users" drop constraint "users_tokenId_fkey",
  add constraint "fk_token_id"
  foreign key ("tokenId")
  references "public"."email_verification_tokens"
  ("id") on update no action on delete no action;
