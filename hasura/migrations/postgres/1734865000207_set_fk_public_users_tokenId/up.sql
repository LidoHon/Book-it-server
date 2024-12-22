alter table "public"."users" drop constraint "fk_token_id",
  add constraint "users_tokenId_fkey"
  foreign key ("tokenId")
  references "public"."email_verification_tokens"
  ("id") on update no action on delete cascade;
